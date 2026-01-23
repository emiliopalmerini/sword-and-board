{
  description = "sword-and-board - TUI character sheet manager for Sword & Board TTRPG";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages = {
          sword-and-board = pkgs.callPackage ./nix/package.nix {};
          default = self.packages.${system}.sword-and-board;
        };

        devShells.default = pkgs.callPackage ./nix/devShell.nix {};
      }
    );
}
