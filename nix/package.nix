{ pkgs }:

pkgs.buildGoModule {
  pname = "sword-and-board";
  version = "0.1.0";
  src = pkgs.lib.cleanSource ../.;

  vendorHash = null; # Will be updated after first build

  subPackages = [ "cmd/sword-and-board" ];

  meta = with pkgs.lib; {
    description = "TUI character sheet manager for Sword & Board TTRPG";
    license = licenses.mit;
    maintainers = [ ];
  };
}
