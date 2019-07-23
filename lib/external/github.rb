require 'base64'
require 'json'
require 'octokit'
require 'yaml'

TAG_PREFIX="cloud.rvu.ontology"

def labels_to_relations(entitiy_id, labels)
  labels.map { |tag, val|
    next if not tag.to_s.start_with? TAG_PREFIX

    type = tag[TAG_PREFIX.length..-1]

    {
      metadata: {
        type: type,
      },
      properties: {
        a: entitiy_id,
        b: val,
      },
    }
   }.compact
end

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
        relations += labels_to_relations("/repository/github.com/#{repo.full_name}", repo_h[:metadata][:tags])
      end

      {
        path: "/repository/github.com/#{repo.full_name}",
        entity: {
          metadata: {
            type: "/entity/v1/repository",
          },
          properties: {
            language: repo[:language],
            license: repo.has_key?(:license) ? repo[:license][:key] : nil,
            created_at: repo[:created_at],
            updated_at: repo[:updated_at],
            pushed_at: repo[:pushed_at],
          },
        },
        relations: relations,
      }
    }
  end
end
