require 'azure_graph_rbac'
require 'json'
require 'ms_rest'

require_relative './utils.rb'

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

    path = "/#{client.tenant_id}/users"
    users = []

    updated_at = DateTime.now.rfc3339

    while path do
      result = client.make_request(:get, path, { query_params: {'api-version' => '1.6'} })

      users += result["value"]

      if result.has_key?("odata.nextLink")
        path = "/#{client.tenant_id}/#{result["odata.nextLink"]}"
      else
        path = nil
      end
    end

    users
      .select { |u| u["accountEnabled"] and u["userType"] == "Member" }
      .map { |u|

      upn = u["userPrincipalName"].downcase
      user, domain = upn.downcase.split("@")

      path = "/person/#{domain}/#{user}"

      alias_entity(
        {
          metadata: {
            type: "/entity/v1/person",
            updated_at: updated_at,
          },
          properties: {
            id: u["objectId"],
            email: u["upn"],
          },
        },
        id: path,
        aliases: [
          "/person/by-email/#{upn}",
          "/person/by-id/#{u["objectId"]}"
        ],
      )
    }.flatten
  end

end
