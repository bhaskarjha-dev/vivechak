{
  description = "Evidence-grounded research for technical decisions — MCP server";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = rec {
          vivechak = pkgs.buildGoModule {
            pname = "vivechak";
            version = "0.1.0";
            src = ./.;
            # To update dependencies when go.mod/go.sum changes, run: nix build .#vivechak 2>&1 | grep "got:"
            vendorHash = "sha256-VVtbgY9jmMmgsiBD0ykRuk4BcG1ihWvlQgcsO4MUu8A=";
            proxyVendor = true;
            ldflags = [ "-s" "-w" "-X github.com/bhaskarjha-dev/vivechak/internal/version.Version=0.1.0" ];
            postInstall = ''
              ln -s $out/bin/vivechak $out/bin/vck
            '';
          };
          vck = vivechak;
          default = vivechak;
        };

        apps = rec {
          vivechak = flake-utils.lib.mkApp { drv = self.packages.${system}.vivechak; };
          vck = flake-utils.lib.mkApp { drv = self.packages.${system}.vck; };
          default = vivechak;
        };
      }
    );
}
