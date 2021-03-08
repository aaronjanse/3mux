{
  inputs.nixpkgs.url = "github:nixos/nixpkgs";
  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forEachSystem = func: nixpkgs.lib.genAttrs systems func;
    in
    {
      defaultApp = forEachSystem (system: {
        type = "app";
        program = "${self.defaultPackage."${system}"}/bin/3mux";
      });
      defaultPackage = forEachSystem (system: nixpkgs.legacyPackages."${system}".callPackage
        ({ lib, buildGoModule, makeWrapper, fetchFromGitHub }:
          buildGoModule {
            name = "3mux-latest";
            src = ./.;
            vendorSha256 = "sha256-tbziQZIA1+b+ZtvA/865c8YQxn+r8HQy6Pqaac2kwcU=";
            excludedPackages = [ "fuzz" ];
            buildInputs = [ makeWrapper ];
            postInstall = ''
              wrapProgram $out/bin/3mux --prefix PATH : $out/bin
            '';
            meta = with lib; {
              description = "Terminal multiplexer inspired by i3";
              longDescription = ''
                3mux is a terminal multiplexer with out-of-the-box support for
                search, mouse-controlled scrollback, and i3-like keybindings.
              '';
              homepage = "https://github.com/aaronjanse/3mux";
              license = licenses.mit;
              platforms = platforms.unix;
            };
          }
        ) { }
      );
    };
}
