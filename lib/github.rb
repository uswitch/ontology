require 'base64'
require 'json'
require 'octokit'

class GitHub

  def sync
    client = Octokit::Client.new(access_token: ENV['GITHUB_TOKEN'])

    client.auto_paginate = true

    client.organization_repositories('uswitch').map { |repo|
      repo_h = repo.to_h

      begin
        metadata_resp = client.contents(repo.full_name, path: ".rvu/metadata")

        repo_h[:metadata] = Base64.decode64(metadata_resp.content)
      rescue Octokit::NotFound
      end

      repo_h
    }
  end
end
