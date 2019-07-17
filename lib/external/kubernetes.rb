require 'k8s-client'

class Kubernetes

  def sync
    client = K8s::Client.autoconfig

    client.api_groups.map { |api|
      api_client = client.api(api)

      api_client.api_resources.map { |resource_client|
        api_client.resource(resource_client.name).list
      }
    }
  end

end
