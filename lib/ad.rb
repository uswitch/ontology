require 'azure_graph_rbac'
require 'json'
require 'ms_rest'

class AD

  def sync
    settings = MsRestAzure::ActiveDirectoryServiceSettings.new
    settings.authentication_endpoint = MsRestAzure::AzureEnvironments::AzureCloud.active_directory_endpoint_url
    settings.token_audience = MsRestAzure::AzureEnvironments::AzureCloud.active_directory_graph_resource_id

    provider = MsRestAzure::ApplicationTokenProvider.new(
      "1181c46f-f004-40a4-959a-2b630e7852df",
      "c7cc8720-1e6c-4925-b1e3-dfabf1a1bd21",
      ENV["AZURE_CLIENT_SECRET"],
      settings,
    )
    credentials = MsRest::TokenCredentials.new(provider)

    client = Azure::GraphRbac::V1_6::GraphRbacClient.new(credentials)
    client.tenant_id = "1181c46f-f004-40a4-959a-2b630e7852df"

    client.users.list.map { |user|
      client.serialize(Azure::GraphRbac::V1_6::Models::User.mapper, user)
    }
  end

end