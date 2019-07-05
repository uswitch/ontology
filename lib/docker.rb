require 'httparty'

#puts response.body, response.code, response.message, response.headers.inspect

def next_link(response)
  next_url = nil
  if response.headers.key? "link"
    match = /\s*<(\/[^>]*)>;\s*rel="next"\s*/.match(response.header["link"])
    if match
      next_url = match[1]
    end
  end

  return next_url
end

class DockerRegistry
  include HTTParty

  base_uri "https://registry.usw.co"

  def sync
    registry.repositories.map { |repo|
      registry.tags(repo).map { |tag|
        manifest = registry.manifest(repo, tag)
        {
          name: "#{repo}:#{tag}",
          created: registry.created_from(manifest),
          labels: registry.labels_from(manifest),
        }
      }
    }.flatten
  end

  def repositories
    list('/v2/_catalog', 'repositories')
  end

  def tags(repo)
    list("/v2/#{repo}/tags/list", 'tags')
  end

  def manifest(repo, ref)
    manifest = self.class.get(
      "/v2/#{repo}/manifests/#{ref}", {
        headers: { "Accept" => "application/vnd.docker.distribution.manifest.v2+json" },
        format: :json,
      })

    config = self.class.get(
      "/v2/#{repo}/blobs/#{manifest["config"]["digest"]}", {
        headers: { "Accept" => "application/vnd.docker.distribution.manifest.v2+json" },
        format: :json,
      })

    config["schemaVersion"] = 2
    config
  end

  def labels_from(manifest)
    case manifest["schemaVersion"]
    when 1
      manifest["history"].map { |layer|
        compat = JSON.parse(layer["v1Compatibility"])
        compat["container_config"]["Labels"]
      }.uniq
    when 2
      cc_l = manifest["container_config"]["Labels"] || {}
      c_l = manifest["config"]["Labels"] || {}

      cc_l.merge(c_l)
    else
      raise "Unknown manifest version: #{manifest["schemaVersion"]}"
    end
  end

  def created_from(manifest)
    case manifest["schemaVersion"]
    when 1
      JSON.parse(manifest["history"][0]["v1Compatibility"])["created"]
    when 2
      manifest["created"]
    else
      raise "Unknown manifest version: #{manifest["schemaVersion"]}"
    end
  end

  private

  def list(path, key)
    list = []

    begin
      response = self.class.get(path)
      list = list + JSON.parse(response.body)[key]
      return list
    end while path = next_link(response)

    list
  end

end