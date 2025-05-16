def generate_tag_args($url):
  .
  | map("-t \($url):\(.)");

generate_tag_args($url) | join(" ")
