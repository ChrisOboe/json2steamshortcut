{
  buildGoModule,
  lib,
  stdenv,
}:
buildGoModule rec {
  name = "json2steamshortcut";
  src = ./src;
  vendorSha256 = null;

  meta = {
    description = "A tool to create a steam shortcut file from a json";
    maintainers = ["chris@oboe.email"];
  };
}
