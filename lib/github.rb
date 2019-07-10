require 'base64'
require 'json'
require 'octokit'
require 'yaml'

TAG_PREFIX="cloud.rvu.ontology"

class GitHub

  def sync
    client = Octokit::Client.new(access_token: ENV['GITHUB_TOKEN'])

    client.auto_paginate = true

    client.organization_repositories('uswitch').map { |repo|
      repo_h = repo.to_h

      begin
        metadata_resp = client.contents(repo.full_name, path: ".rvu/metadata")

        repo_h[:metadata] = YAML.load(Base64.decode64(metadata_resp.content))
      rescue Octokit::NotFound
      end

      relations = []

      if repo_h.has_key?(:metadata) and repo_h[:metadata].has_key?(:tags)
        repo_h[:metadata][:tags].each { |tag, val|
          return if not tag.starts_with TAG_PREFIX

          type = tag[TAG_PREFIX.length..-1]
          relations.push(
            {
              metadata: {
                type: type,
              },
              properties: {
                other: val
              },
            }
          )
        }

        puts relations
      end

      {
        path: "/repository/github.com/#{repo.full_name}.json",
        entity: {
          metadata: {
            type: "/entities/v1/repository",
          },
          relations: relations,
          properties: {
            language: repo[:language],
            license: repo.has_key?(:license) ? repo[:license][:key] : nil,
            created_at: repo[:created_at],
            updated_at: repo[:updated_at],
            pushed_at: repo[:pushed_at],
          },
        },
      }
    }
  end
end
