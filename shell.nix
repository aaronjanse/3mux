with import <nixpkgs>{};
stdenv.mkDerivation rec {
    name = "multiplexer";
    buildInputs =  [ autoreconfHook pkgconfig cmake ncurses go libtsm ];
    shellHook = ''
        GODEBUG=cgocheck=0 go run *.go
        exit
    '';

    GOPATH="/home/ajanse/.go";
}
