require 'httparty'
require 'ruby-progressbar'
require 'uri'

require_relative './utils.rb'

#puts response.body, response.code, response.message, response.headers.inspect

# stolen from github.com/distribution/reference
REFERENCE_REGEX = /^((?:(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])(?:(?:\.(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]))+)?(?::[0-9]+)?\/)?[a-z0-9]+(?:(?:(?:[._]|__|[-]*)[a-z0-9]+)+)?(?:(?:\/[a-z0-9]+(?:(?:(?:[._]|__|[-]*)[a-z0-9]+)+)?)+)?)(?::([\w][\w.-]{0,127}))?(?:@([A-Za-z][A-Za-z0-9]*(?:[-_+.][A-Za-z][A-Za-z0-9]*)*[:][[:xdigit:]]{32,}))?$/
DIGEST_REGEX = /[A-Za-z][A-Za-z0-9]*(?:[-_+.][A-Za-z][A-Za-z0-9]*)*[:][[:xdigit:]]{32,}/

DIGEST_LENGTHS = {
  "sha256" => 64,
  "sha512" => 128,
}

def parse_image_reference(reference)
  raise "ErrNameEmpty" if reference.length == 0

  ref_match = REFERENCE_REGEX.match(reference)

  raise "ErrReferenceInvalidFormat" if not ref_match

  domain_uri = URI("http://" + ref_match[1])

  if ref_match[0].include?("@")
    digest_match = DIGEST_REGEX.match(ref_match[0].split("@")[1])

    if digest_match
      digest_type = digest_match[0].split(":")[0]
      if DIGEST_LENGTHS.has_key?(digest_type)
        if digest_match[0].split(":")[-1].length != DIGEST_LENGTHS[digest_type]
          raise "digest.ErrDigestInvalidLength"
        end
      else
        raise "digest.ErrDigestUnsupported"
      end
    else
      raise "invalid digest"
    end
  end

  raise "ErrNameTooLong" if ref_match[1].length > 255

  {
    repository: ref_match[1],
    domain: domain_uri.host + ((domain_uri.port && domain_uri.port != 80) ? ":#{domain_uri.port}" : ""),
    tag: (ref_match[2] == nil or ref_match[2].length == 0) ? "latest" : ref_match[2],
    digest: digest_match ? digest_match[0] : nil,
  }
end

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

HUB_DOMAIN="registry-1.docker.io"

module Ontology

  module Source

    class DockerRegistry
      include HTTParty

      def initialize(domain, token=nil)
        prefix = "#{domain}/"
        if domain == nil
          domain = HUB_DOMAIN
          prefix = ""
        end
        @domain = domain
        @prefix = prefix

        @headers = {}
        if token
          @headers["Authorization"] = "Bearer #{token}"
        end

        self.class.base_uri "https://#{domain}"
      end

      def image(repo, tag)
        begin
          digest, manifest = manifest(repo, tag)

          path = "/image/container/#{@prefix}#{repo}/#{digest}"
          labels = labels_from(manifest)
          relations = []

          updated_at = DateTime.now.rfc3339

          if labels.has_key?("org.label-schema.vcs-url")
            vcs_uri = URI(labels["org.label-schema.vcs-url"])

            without_ext = File.join(File.dirname(vcs_uri.path), File.basename(vcs_uri.path, ".*"))

            relations << {
              metadata: {
                type: "/relation/v1/was_built_by",
                updated_at: updated_at,
              },
              properties: {
                a: path,
                b: "/repository/#{vcs_uri.host}#{without_ext}",
                ref: labels["org.label-schema.vcs-ref"],
                at: labels["org.label-schema.build-date"],
              }
            }
          end

          alias_entity(
            {
              metadata: {
                type: "/entity/v1/image/container",
                updated_at: updated_at,
              },
              properties: {
                server: @domain,
                repository: repo,
                digest: digest,
                created_at: created_from(manifest),
              },
            },
            id: path,
            aliases: [
              "/image/container/#{@prefix}#{repo}/#{tag}",
            ],
          ) + add_ids_to(relations, base: path)
        rescue StandardError => e
          $stderr.puts manifest
          $stderr.puts e.message
          $stderr.puts e.backtrace.inspect
          []
        end
      end

      def sync
        puts "Loading all repositories"
        Parallel.map(repositories, progress: "Getting all repositories", in_processes: 20) { |repo|
          Parallel.map(tags(repo)) { |tag|
            image(repo, tag)
          }
        }.flatten
      end

      def repositories
        list('/v2/_catalog?n=1000000', 'repositories')
      end

      def tags(repo)
        list("/v2/#{repo}/tags/list?n=1000000", 'tags')
      end

      def manifest(repo, ref)
        manifest = self.class.get(
          "/v2/#{repo}/manifests/#{ref}", {
            headers: @headers.merge({ "Accept" => "application/vnd.docker.distribution.manifest.v2+json, application/vnd.docker.distribution.manifest.v1+json" }),
            format: :json,
          })

        if manifest["config"]
          manifest_digest = "sha256:#{Digest::SHA256.hexdigest manifest.body}"
        else
          #$stderr.puts "WARN: v2.1 manifest, cannot calculate digest so relying on server"
          manifest_digest = manifest.header["Docker-Content-Digest"]
        end

        if manifest_digest != manifest.header["Docker-Content-Digest"]
          puts "[#{repo}:#{ref}] #{manifest.header["Docker-Content-Digest"]} != #{manifest_digest}"
        end

        if manifest["config"]
          config = self.class.get(
            "/v2/#{repo}/blobs/#{manifest["config"]["digest"]}", {
              headers: @headers.merge({ "Accept" => "application/vnd.docker.distribution.manifest.v2+json" }),
              format: :json,
            })

          config["schemaVersion"] = 2
        else
          config = manifest
        end

        return manifest_digest, config
      end

      def labels_from(manifest)
        case manifest["schemaVersion"]
        when 1
          manifest["history"].map { |layer|
            compat = JSON.parse(layer["v1Compatibility"])
            compat["container_config"]["Labels"]
          }.compact.reduce(&:merge)
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
        rescue StandardError => e
          $stderr.puts e.message
          $stderr.puts e.backtrace.inspect
        end while path = next_link(response)

        list
      end

    end

  end

end
