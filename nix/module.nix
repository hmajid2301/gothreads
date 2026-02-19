{ config, pkgs, lib, ... }:

let
  cfg = config.services.gothreads;
  stateDir = "/var/lib/gothreads";
in
{
  meta.maintainers = with lib.maintainers; [ ];

  options.services.gothreads = {
    enable = lib.mkEnableOption "GoThreads, a privacy-first self-hosted digital wardrobe manager";

    package = lib.mkPackageOption pkgs "gothreads" { };

    address = lib.mkOption {
      type = lib.types.str;
      default = "localhost";
      description = "Web interface address.";
    };

    port = lib.mkOption {
      type = lib.types.port;
      default = 8556;
      description = "Web interface port.";
    };

    openFirewall = lib.mkOption {
      type = lib.types.bool;
      default = false;
      description = "Open the firewall for the GoThreads port.";
    };

    user = lib.mkOption {
      type = lib.types.str;
      default = "gothreads";
      description = "User account under which GoThreads runs.";
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = "gothreads";
      description = "Group under which GoThreads runs.";
    };

    extraConfig = lib.mkOption {
      type = lib.types.attrs;
      default = { };
      description = ''
        Extra environment variables for GoThreads.
        
        Common options:
        - JWT_SECRET: Secret key for JWT tokens (required for auth)
        - SKIP_AUTH: Set to "true" to disable authentication
        - DATABASE_URL: PostgreSQL connection string
        - OLLAMA_URL: Ollama API URL (default: http://localhost:11434)
        - REMBG_URL: Rembg service URL (default: http://localhost:5000)
        - VTON_URL: Virtual try-on service URL
        - S3_ENDPOINT: S3 endpoint URL
        - S3_BUCKET: S3 bucket name
        - S3_ACCESS_KEY: S3 access key
        - S3_SECRET_KEY: S3 secret key
      '';
      example = {
        SKIP_AUTH = "true";
        OLLAMA_URL = "http://localhost:11434";
        REMBG_URL = "http://localhost:5000";
      };
    };

    database.createLocally = lib.mkOption {
      type = lib.types.bool;
      default = false;
      description = "Configure local PostgreSQL database for GoThreads.";
    };

    ai = {
      ollama = lib.mkOption {
        type = lib.types.bool;
        default = false;
        description = "Enable local Ollama service for AI features.";
      };

      rembg = lib.mkOption {
        type = lib.types.bool;
        default = false;
        description = "Enable local Rembg service for background removal.";
      };

      acceleration = lib.mkOption {
        type = lib.types.nullOr (lib.types.enum [ "cuda" "rocm" ]);
        default = null;
        description = "GPU acceleration for AI services.";
      };
    };
  };

  config = lib.mkIf cfg.enable {
    users.users = lib.mkIf (cfg.user == "gothreads") {
      gothreads = {
        inherit (cfg) group;
        isSystemUser = true;
        home = stateDir;
      };
    };

    users.groups = lib.mkIf (cfg.group == "gothreads") {
      gothreads = { };
    };

    systemd.services.gothreads = {
      description = "GoThreads Digital Wardrobe Manager";
      requires = lib.optional cfg.database.createLocally "postgresql.target"
                 ++ lib.optional cfg.ai.ollama "ollama.service";
      after = lib.optional cfg.database.createLocally "postgresql.target"
              ++ lib.optional cfg.ai.ollama "ollama.service";

      serviceConfig = {
        ExecStart = "${cfg.package}/bin/api";
        Restart = "on-failure";
        RestartSec = "5s";

        User = cfg.user;
        Group = cfg.group;
        StateDirectory = "gothreads";
        WorkingDirectory = stateDir;

        BindReadOnlyPaths = [
          "${config.security.pki.caBundle}:/etc/ssl/certs/ca-certificates.crt"
          builtins.storeDir
          "-/etc/resolv.conf"
          "-/etc/nsswitch.conf"
          "-/etc/hosts"
          "-/etc/localtime"
        ] ++ lib.optional cfg.database.createLocally "/run/postgresql";

        CapabilityBoundingSet = "";
        LockPersonality = true;
        MemoryDenyWriteExecute = true;
        NoNewPrivileges = true;
        PrivateDevices = true;
        PrivateUsers = true;
        ProtectClock = true;
        ProtectControlGroups = true;
        ProtectHome = true;
        ProtectHostname = true;
        ProtectKernelLogs = true;
        ProtectKernelModules = true;
        ProtectKernelTunables = true;
        ProtectProc = "invisible";
        ProtectSystem = "strict";
        RestrictAddressFamilies = [ "AF_UNIX" "AF_INET" "AF_INET6" ];
        RestrictNamespaces = true;
        RestrictRealtime = true;
        RestrictSUIDSGID = true;
        SystemCallArchitectures = "native";
        SystemCallFilter = [ "@system-service" "~@privileged" ];
        UMask = "0066";
      };

      environment = {
        PORT = toString cfg.port;
        GOTHREADS_CONFIG = "${stateDir}/config.yaml";
      }
      // lib.optionalAttrs cfg.database.createLocally {
        DATABASE_URL = "postgres://gothreads@/gothreads?host=/run/postgresql";
      }
      // lib.optionalAttrs cfg.ai.ollama {
        OLLAMA_URL = "http://localhost:11434";
      }
      // lib.optionalAttrs cfg.ai.rembg {
        REMBG_URL = "http://localhost:5000";
      }
      // (lib.mapAttrs (_: toString) cfg.extraConfig);

      wantedBy = [ "multi-user.target" ];

      preStart = ''
        if [ ! -f ${stateDir}/config.yaml ]; then
          cat > ${stateDir}/config.yaml << 'EOF'
        server:
          port: ${toString cfg.port}
          host: ${cfg.address}
        
        ai:
          provider: local
          ollama:
            url: http://localhost:11434
            vision_model: llava:13b
            text_model: llama3.2:3b
        
        image_processing:
          rembg_url: http://localhost:5000
          auto_remove_bg: true
        
        storage:
          s3:
            enabled: false
            endpoint: http://localhost:9000
            bucket: gothreads
        
        weather:
          location: London
          unit: celsius
        EOF
        fi
      '';
    };

    services.postgresql = lib.mkIf cfg.database.createLocally {
      enable = true;
      ensureDatabases = [ "gothreads" ];
      ensureUsers = [{ name = "gothreads"; ensureDBOwnership = true; }];
    };

    services.ollama = lib.mkIf cfg.ai.ollama {
      enable = true;
      acceleration = cfg.ai.acceleration;
    };

    virtualisation.oci-containers.containers.rembg = lib.mkIf cfg.ai.rembg {
      image = "danielgatis/rembg:latest";
      ports = [ "5000:5000" ];
      cmd = [ "p" "--host" "0.0.0.0" "--port" "5000" ];
    };

    networking.firewall.allowedTCPPorts = lib.mkIf cfg.openFirewall [ cfg.port ];
  };
}
