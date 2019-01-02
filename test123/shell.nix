with import <nixpkgs>{};
stdenv.mkDerivation rec {
    name = "LED";
    buildInputs =  [ autoreconfHook pkgconfig cmake go libtsm ];
    shellHook = ''
        go run *.go
        exit
    '';

    GOPATH="/home/ajanse/.go";
}
