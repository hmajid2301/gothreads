{ pkgs, system }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gopls
    gotools
    go-tools
    golangci-lint
    postgresql
    minio
    curl
    jq
    gnumake
    pkg-config
  ];

  shellHook = ''
    export GOPATH=$HOME/go
    export PATH=$GOPATH/bin:$PATH
    export PORT=8556
    export DATABASE_URL="postgres://postgres:postgres@localhost:5432/gothreads?sslmode=disable"
    export S3_ENDPOINT="http://localhost:9000"
    export S3_BUCKET="gothreads"
    export S3_ACCESS_KEY="minioadmin"
    export S3_SECRET_KEY="minioadmin"
    export OLLAMA_URL="http://localhost:11434"
    export REMBG_URL="http://localhost:5000"

    echo ""
    echo "🧵 GoThreads Dev Environment"
    echo ""
    echo "  cd mockups/api && go run ."
    echo ""
  '';
}
