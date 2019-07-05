require "google/cloud/asset"
require "google/cloud/storage"

class GCP

  def sync

    asset_client = Google::Cloud::Asset.new(version: :v1)
    storage_client = Google::Cloud::Storage.new(project_id: "ontology-243710")

    bucket_name = "rvu-ontology-gcp-assets"
    file_name = "assets.json"

    parent = "folders/940373838158"
    output_config = {
      gcs_destination: { uri: "gs://#{bucket_name}/#{file_name}", },
    }

    out = ""

    operation = asset_client.export_assets(parent, output_config, content_type: Google::Cloud::Asset::V1::ContentType::RESOURCE)
    operation.on_done {
      bucket = storage_client.bucket(bucket_name)
      file = bucket.file(file_name)

      downloaded = file.download

      downloaded.rewind
      out = downloaded.read
    }

    operation.wait_until_done!

    out
  end

end
