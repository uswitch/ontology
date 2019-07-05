require 'aws-sdk-configservice'
require 'json'

class AWS

  def sync
    client = Aws::ConfigService::Client.new

    next_token = nil
    all_resources = []
    begin
      resources, next_token = find_resources(client, next_token)
      all_resources = all_resources + resources
    end while next_token != ""

    all_resources
  end

  def find_resources(client, next_token=nil)
    resp = client.select_resource_config(
      expression: "SELECT *, configuration, tags",
      next_token: next_token,
      limit: 100,
    )
    return resp[:results].map { |r| JSON.parse(r) }, resp[:next_token]
  end

end
