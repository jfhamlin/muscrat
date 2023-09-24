{ sources ? import ./nix/sources.nix
, pkgs ? import sources.nixpkgs {}
, frameworks ? pkgs.darwin.apple_sdk.frameworks
}:

pkgs.mkShell {
  nativeBuildInputs = [
    pkgs.go_1_21
    pkgs.nodejs

    pkgs.SDL2
    pkgs.portaudio

    pkgs.pkgconfig

    frameworks.Security
    frameworks.Cocoa
    frameworks.WebKit
    frameworks.UniformTypeIdentifiers
    frameworks.ForceFeedback

    # for fyne
    frameworks.UserNotifications
  ];
}
