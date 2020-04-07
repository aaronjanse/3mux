with import <nixpkgs>{};
stdenv.mkDerivation rec {
    name = "multiplexer";
    buildInputs =  [ autoreconfHook pkgconfig cmake ncurses go libtsm ];

    GOPATH="/home/ajanse/.go";
    GODEBUG="cgocheck=0";
}
