{
  description = "GoThreads - Privacy-first self-hosted digital wardrobe manager";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";

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
      ...
    }:
    (flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ gomod2nix.overlays.default ];
        };

        gothreads = import ./nix/package.nix {
          inherit pkgs;
          src = ./mockups/api;
          root = ./.;
        };

        docker = import ./nix/docker.nix {
          inherit pkgs;
          gothreads = gothreads;
        };

        devShell = import ./nix/shell.nix {
          inherit pkgs system;
        };
      in
      {
        packages = {
          default = gothreads;
          inherit gothreads docker;
        };

        devShells.default = devShell;

        checks.build = gothreads;
      }
    ))
    // {
      nixosModules.default = import ./nix/module.nix;
    };
}
