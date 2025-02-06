{
  description = "Go development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }: 
    let 
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
    in {
      devShells.${system}.default = pkgs.mkShell {
        buildInputs = [
          pkgs.go
          pkgs.gopls  # Go language server
          pkgs.golangci-lint # Linter
          pkgs.delve  # Debugger
          pkgs.fish   # Fish shell
        ];

        shellHook = ''
          echo "Go development environment is ready!"
          export GOPATH=$PWD/go
          export PATH=$GOPATH/bin:$PATH
          exec fish
        '';
      };
    };
}
