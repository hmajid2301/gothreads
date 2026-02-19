{ pkgs, gothreads }:

pkgs.dockerTools.buildLayeredImage {
  name = "gothreads/gothreads";
  tag = "latest";

  contents = [
    gothreads
    pkgs.cacert
    pkgs.bashInteractive
  ];

  config = {
    Cmd = [ "/bin/api" ];
    WorkingDir = "/share/gothreads";
    ExposedPorts = {
      "8556/tcp" = { };
    };
    Env = [
      "SSL_CERT_FILE=/etc/ssl/certs/ca-bundle.crt"
      "GOTHREADS_CONFIG=/share/gothreads/config.yaml"
    ];
    Labels = {
      "org.opencontainers.image.title" = "GoThreads";
      "org.opencontainers.image.description" = "Privacy-first digital wardrobe manager";
      "org.opencontainers.image.version" = "1.0.0";
    };
  };
}
