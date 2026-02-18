{
  description = "go-threads: Privacy-first self-hosted digital wardrobe manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    playwright.url = "github:pietdevries94/playwright-web-flake/1.57.0";

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      gomod2nix,
      playwright,
      ...
    }:
    (flake-utils.lib.eachDefaultSystem (
      system:
      let
        overlay = final: prev: {
          inherit (playwright.packages.${system})
            playwright-test
            playwright-driver
            ;
        };
        pkgs = import nixpkgs {
          inherit system;
          overlays = [
            gomod2nix.overlays.default
            overlay
          ];
        };

        # Backend Go packages
        backendPackages = with pkgs; [
          go
          goose
          air
          golangci-lint
          gotools
          gotestsum
          gocover-cobertura
          go-task
          go-mockery
          templ
          sqlc
          sqlfluff
        ];

        # Frontend packages (Bun + Svelte for outfit canvas only)
        frontendPackages = with pkgs; [
          bun
          tailwindcss_4
        ];

        # Testing packages
        testPackages = with pkgs; [
          playwright-driver
          gotestsum
        ];

        # Development tools
        devPackages = with pkgs; [
          watchman
          concurrently
          rustywind
          direnv
        ];

        # AI/LLM packages (optional - can use Docker Ollama instead)
        aiPackages = with pkgs; [
          ollama
          # llama-cpp # Uncomment for llama.cpp CLI
        ];

        allPackages = backendPackages ++ frontendPackages ++ testPackages ++ devPackages ++ aiPackages;

        devShellPackages =
          allPackages
          ++ [
            gomod2nix.packages.${system}.default
          ];
      in
      rec {
        # Backend Go application
        packages.backend = pkgs.buildGoApplication {
          pname = "gothreads";
          version = "0.1.0";
          src = ./backend;
          modules = ./backend/gomod2nix.toml;
          pwd = ./backend;

          # Build frontend assets and generate code before building Go binary
          preBuild = ''
            # Generate Templ files
            ${pkgs.templ}/bin/templ generate

            # Build Tailwind CSS
            ${pkgs.tailwindcss}/bin/tailwindcss -i ./static/css/input.css -o ./static/css/output.css --minify

            # Build Svelte component (if frontend exists)
            if [ -d ../frontend ]; then
              cd ../frontend
              ${pkgs.bun}/bin/bun install --frozen-lockfile
              ${pkgs.bun}/bin/bun run build
              cd ../backend
            fi
          '';
        };

        # Default package is just the backend (frontend is embedded static files)
        packages.default = packages.backend;

        # Docker container
        packages.container = pkgs.dockerTools.buildImage {
          name = "gothreads";
          tag = "latest";
          created = "now";
          copyToRoot = pkgs.buildEnv {
            name = "image-root";
            paths = [
              packages.default
              pkgs.cacert
            ];
            pathsToLink = [ "/bin" ];
          };
          config = {
            ExposedPorts = {
              "8080/tcp" = { };
            };
            Cmd = [ "${packages.backend}/bin/gothreads" ];
            Env = [
              "SSL_CERT_FILE=${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
              "SSL_CERT_DIR=${pkgs.cacert}/etc/ssl/certs/"
            ];
            Labels = {
              service = "gothreads";
              version = "0.1.0";
            };
          };
        };

        # Development shell
        devShells.default = pkgs.mkShell {
          packages = devShellPackages;

          shellHook = ''
            # Playwright configuration
            export PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
            export PLAYWRIGHT_BROWSERS_PATH="${pkgs.playwright-driver.browsers}"
            export PLAYWRIGHT_NODEJS_PATH="${pkgs.nodejs}/bin/node"
            export PLAYWRIGHT_DRIVER_PATH="${pkgs.playwright-driver}"

            # Goose migration configuration
            export GOOSE_DRIVER=postgres
            export GOOSE_MIGRATION_DIR="backend/store/db/sqlc/migrations"

            # Go build configuration
            export CGO_ENABLED=0

            # Application configuration
            export GOTHREADS_LOG_LEVEL=debug
            export GOTHREADS_ENVIRONMENT=local
            export GOTHREADS_WEBSERVER_HOST=0.0.0.0
            export GOTHREADS_WEBSERVER_PORT=8080

            # Database configuration
            export GOTHREADS_DB_DATABASE_URL="postgres://postgres:postgres@localhost:15433/gothreads?sslmode=disable"

            # OAuth configuration (mock-oauth2-server)
            export GOTHREADS_OAUTH_SKIP_AUTH=true
            export GOTHREADS_OAUTH_JWKS_URL=http://localhost:8090/default/jwks
            export GOTHREADS_OAUTH_CLIENT_ID=debugger
            export GOTHREADS_OAUTH_CLIENT_SECRET=secret
            export GOTHREADS_OAUTH_AUTHORIZE_URL=http://localhost:8090/default/authorize
            export GOTHREADS_OAUTH_TOKEN_URL=http://localhost:8090/default/token
            export GOTHREADS_OAUTH_REDIRECT_URL=http://localhost:8080/callback

            # S3 / SeaweedFS configuration
            export GOTHREADS_S3_ENDPOINT=http://localhost:8333
            export GOTHREADS_S3_BUCKET=gothreads
            export GOTHREADS_S3_ACCESS_KEY=admin
            export GOTHREADS_S3_SECRET_KEY=admin123

            # AI Worker configuration
            export AI_WORKER_CONCURRENCY=3
            export ENABLE_GPU=false

            # Ollama configuration (local LLM)
            export GOTHREADS_OLLAMA_URL=http://localhost:11434
            export GOTHREADS_OLLAMA_VISION_MODEL=llava:7b
            export GOTHREADS_OLLAMA_TEXT_MODEL=llama3.2:3b

            # VTON service URLs
            export GOTHREADS_LADIVTON_SHOES_URL=http://localhost:8558

            # Only show welcome message once per shell session
            if [ -z "$GOTHREADS_SHELL_INITIALIZED" ]; then
              export GOTHREADS_SHELL_INITIALIZED=1
              echo "🧵 go-threads development environment loaded"
              echo "📁 Backend: ./backend"
              echo "📁 Frontend: ./frontend"
              echo "🚀 Run 'task dev' to start development servers"
            fi
          '';
        };
      }
    ))
    // {
      # NixOS Module
      nixosModules.default = { config, lib, pkgs, ... }:
        with lib;
        let
          cfg = config.services.gothreads;
        in {
          options.services.gothreads = {
            enable = mkEnableOption "GoThreads digital wardrobe manager";
            
            port = mkOption {
              type = types.int;
              default = 8556;
              description = "Port to listen on";
            };
            
            dataDir = mkOption {
              type = types.path;
              default = "/var/lib/gothreads";
              description = "Data directory for GoThreads";
            };
            
            configFile = mkOption {
              type = types.nullOr types.path;
              default = null;
              description = "Path to config.yaml";
            };
            
            databaseUrl = mkOption {
              type = types.str;
              default = "postgres:///gothreads";
              description = "PostgreSQL connection URL";
            };
            
            aiProvider = mkOption {
              type = types.enum [ "local" "cloud" ];
              default = "local";
              description = "AI provider to use";
            };
            
            ollamaUrl = mkOption {
              type = types.str;
              default = "http://localhost:11434";
              description = "Ollama API URL";
            };
            
            s3Endpoint = mkOption {
              type = types.str;
              default = "http://localhost:8333";
              description = "S3-compatible storage endpoint";
            };
            
            s3Bucket = mkOption {
              type = types.str;
              default = "gothreads";
              description = "S3 bucket name";
            };
          };
          
          config = mkIf cfg.enable {
            users.users.gothreads = {
              isSystemUser = true;
              group = "gothreads";
              home = cfg.dataDir;
              createHome = true;
            };
            users.groups.gothreads = {};
            
            systemd.services.gothreads = {
              description = "GoThreads Digital Wardrobe Manager";
              wantedBy = [ "multi-user.target" ];
              after = [ "network.target" "postgresql.service" ];
              wants = [ "postgresql.service" ];
              
              environment = {
                PORT = toString cfg.port;
                DATABASE_URL = cfg.databaseUrl;
                AI_PROVIDER = cfg.aiProvider;
                OLLAMA_URL = cfg.ollamaUrl;
                S3_ENDPOINT = cfg.s3Endpoint;
                S3_BUCKET = cfg.s3Bucket;
              } // optionalAttrs (cfg.configFile != null) {
                GOTHREADS_CONFIG = cfg.configFile;
              };
              
              serviceConfig = {
                Type = "simple";
                User = "gothreads";
                Group = "gothreads";
                WorkingDirectory = cfg.dataDir;
                ExecStart = "${self.packages.${pkgs.system}.default}/bin/gothreads";
                Restart = "on-failure";
                RestartSec = "5s";
                NoNewPrivileges = true;
                ProtectSystem = "strict";
                ProtectHome = true;
                PrivateTmp = true;
                ReadWritePaths = [ cfg.dataDir ];
              };
            };
            
            networking.firewall.allowedTCPPorts = [ cfg.port ];
          };
        };
    };
}
