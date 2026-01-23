{ pkgs }:

pkgs.mkShell {
  packages = [ pkgs.go_1_23 ];
}
