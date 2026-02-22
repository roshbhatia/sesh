{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    git
    # Add your development tools here
    # Examples:
    # nodejs
    # go
    # python3
    # rustc
  ];
  
  shellHook = ''
    echo "Session: $(basename $PWD)"
  '';
}
