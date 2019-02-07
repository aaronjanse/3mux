with import <nixpkgs>{};
stdenv.mkDerivation rec {
    name = "vscode";
    buildInputs =  [ autoreconfHook pkgconfig cmake ncurses go libtsm ];
    shellHook = ''
        code .
        exit
    '';

    GOPATH="/home/ajanse/.go";
}
