{buildGoModule, ...}:
buildGoModule {
  name = "json2steamshortcut";
  src = ./src;
  vendorHash = null;

  meta = {
    description = "A tool to create a steam shortcut file from a json";
    mainProgram = "json2steamshortcut";
    maintainers = ["chris@oboe.email"];
  };
}
