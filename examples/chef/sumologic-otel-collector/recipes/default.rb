sumologic_otel_collector 'sumologic-otel-collector' do
  install_token "dummy"
  collector_tags ({"abc" => "def"})
end
