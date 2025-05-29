def matches_target($key; $target):
  $key | test("^manifest_\($target)_");

def filter_target($target):
  .
  | to_entries
  | map(select(matches_target(.key; $target)));

def target_source:
  {
    image: .["image.name"],
    digest: .["containerimage.digest"]
  };

def target_source_as_string:
  target_source | "\(.image)@\(.digest)";

def target_sources($target):
  .
  | filter_target($target)
  | map(.value | fromjson)
  | map(target_source_as_string);

target_sources($target) | join(" ")
