def generate_tag_args($url):
  .
  | map("\($url):\(.)");

generate_tag_args($url) | join(" ")
