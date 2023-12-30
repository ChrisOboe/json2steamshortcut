{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};
    in rec {
      packages = rec {
        default = json2steamshortcut;
        json2steamshortcut = pkgs.callPackage ./default.nix {};
      };

      checks = let
        simpleCheck = {
          name,
          check,
        }:
          pkgs.stdenvNoCC.mkDerivation {
            inherit name;
            dontBuild = true;
            doCheck = true;
            installPhase = "touch $out";
            src = ./.;
            checkPhase = check;
          };
      in {
        lint = simpleCheck {
          name = "lint";
          check = ''
            #!${pkgs.bash}/bin/bash

            set -x

            export PATH=$PATH:${pkgs.go}/bin
            export GOLANGCI_LINT_CACHE="$PWD/cache"
            export XDG_CACHE_HOME=$GOLANGCI_LINT_CACHE

            ${pkgs.golangci-lint}/bin/golangci-lint run --timeout=10m --tests=false
          '';
        };
      };

      devShells.default = pkgs.mkShell {
        buildInputs = [
          # dev
          pkgs.go

          # linter
          pkgs.golangci-lint
          pkgs.statix

          # formatter
          pkgs.gofumpt
          pkgs.alejandra

          # ide
          pkgs.gopls
        ];
      };
    });
}
