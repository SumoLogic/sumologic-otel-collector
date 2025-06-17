def target_source:
  {
    image: .["image.name"],
    digest: .["containerimage.digest"]
  };

def target_source_as_string:
  target_source | "\(.image)@\(.digest)\n";

target_source_as_string
