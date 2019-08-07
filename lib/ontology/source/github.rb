require 'base64'
require 'json'
require 'octokit'
require 'yaml'

module Ontology

  module Source

    class GitHub

      def sync
        client = Octokit::Client.new(access_token: ENV['GITHUB_TOKEN'])

        client.auto_paginate = true

        client.organization_repositories('uswitch').map { |repo|
          repo_h = repo.to_h

          updated_at = DateTime.now.rfc3339

          begin
            metadata_resp = client.contents(repo.full_name, path: ".github/metadata")

            repo_h[:metadata] = YAML.load(Base64.decode64(metadata_resp.content))
          rescue Octokit::NotFound
          end

          id = "/repository/github.com/#{repo.full_name}"

          relations = []

          if repo_h.has_key?(:metadata) and repo_h[:metadata].has_key?(:tags)
            relations += labels_to_relations(id, updated_at, repo_h[:metadata][:tags])
          end

          [
            {
              metadata: {
                id: id,
                type: "/entity/v1/repository",
                updated_at: updated_at,
              },
              properties: {
                server: "github.com",
                path: repo.full_name,
                archvied: repo[:archived],
                disabled: repo[:disabled],
                language: repo[:language],
                license: repo.has_key?(:license) ? repo[:license][:key] : nil,
                created_at: repo[:created_at],
                updated_at: repo[:updated_at],
                pushed_at: repo[:pushed_at],
              },
            },
          ] + add_ids_to(relations, base: id)
        }.flatten
      end
    end

  end

end
