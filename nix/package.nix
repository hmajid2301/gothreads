{ pkgs, src, root }:

pkgs.buildGoApplication {
  pname = "gothreads";
  version = "1.0.0";

  src = src;
  modules = src + "/gomod2nix.toml";

  ldflags = [
    "-s"
    "-w"
    "-X main.version=1.0.0"
  ];

  postInstall = ''
    mkdir -p $out/share/gothreads
    cp -r ${root}/mockups/css $out/share/gothreads/
    cp -r ${root}/mockups/js $out/share/gothreads/
    cp -r ${root}/mockups/images $out/share/gothreads/ || true
    cp ${root}/mockups/*.html $out/share/gothreads/
    cp ${root}/config.yaml $out/share/gothreads/ || true

    mkdir -p $out/bin
    cat > $out/bin/gothreads << 'EOF'
#!/bin/sh
export GOTHREADS_CONFIG="''${GOTHREADS_CONFIG:-$out/share/gothreads/config.yaml}"
cd $out/share/gothreads
exec $out/bin/api "$@"
EOF
    chmod +x $out/bin/gothreads
  '';
}
