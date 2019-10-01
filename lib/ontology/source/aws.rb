require 'aws-sdk-configservice'
require 'aws-sdk-ec2'
require 'aws-sdk-elasticsearchservice'
require 'aws-sdk-elasticache'
require 'aws-sdk-efs'
require 'json'
require 'resolv'

def tag_relations(a, updated_at, resource)
  resource["tags"].map { |tag|
    next if not tag["key"].start_with? TAG_PREFIX

    type = tag["key"][TAG_PREFIX.length..-1]

    {
      metadata: {
        type: type,
      },
      properties: {
        a: a,
        b: tag["value"],
      }
    }
  }.compact
end

def eip_entity(resource)
  path = "/ip_v4_address/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/#{resource["resourceId"]}"
  updated_at = DateTime.iso8601(resource["configurationItemCaptureTime"]).rfc3339
  relations = tag_relations(path, updated_at, resource)

  if resource["configuration"]["networkInterfaceId"]
    relations << {
      metadata: {
        type: "/relation/v1/is_part_of",
        updated_at: updated_at,
      },
      properties: {
        a: path,
        b: "/network_interface/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/#{resource["configuration"]["networkInterfaceId"]}",
      },
    }
  elsif resource["configuration"]["instanceId"]
    relations << {
      metadata: {
        type: "/relation/v1/is_part_of",
        updated_at: updated_at,
      },
      properties: {
        a: path,
        b: "/computer/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/#{resource["configuration"]["instanceId"]}",
      },
    }
  end

  [
    {
      metadata: {
        id: path,
        type: "/entity/v1/ip_v4_address",
        updated_at: updated_at,
      },
      properties: {
        provider: "aws",
        address: resource["resourceName"],
      }
    },
  ] + add_ids_to(relations, base: path)
end

def instance_entity(resource)
  path = "/computer/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/#{resource["resourceId"]}"
  updated_at = DateTime.iso8601(resource["configurationItemCaptureTime"]).rfc3339

  # we should get the user data and then parse the containers that are run

  relations = tag_relations(path, updated_at, resource)

  [
    {
      metadata: {
        id: path,
        type: "/entity/v1/computer",
        updated_at: updated_at,
      },
      properties: {
        provider: "aws",
        image: resource["configuration"]["imageId"],
      }
    }
  ] + add_ids_to(relations, base: path)
end

def eni_entity(resource)
  path = "/network_interface/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/#{resource["resourceId"]}"
  updated_at = DateTime.iso8601(resource["configurationItemCaptureTime"]).rfc3339

  ip_addresses = resource["configuration"]["privateIpAddresses"].map { |entry|
    entry["privateIpAddress"]
  }

  if resource["configuration"].has_key?("association")
    ip_addresses << resource["configuration"]["association"]["publicIp"]
  end

  symlinks = ip_addresses.map { |ip_address|
    "/network_interface/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/by-ip/#{ip_address}"
  }

  relations = tag_relations(path, updated_at, resource)
  description = resource["configuration"]["description"]

  interface_type = resource["configuration"]["interfaceType"]
  requester_id = resource["configuration"]["requesterId"]

  if resource["configuration"]["status"] == "in-use"
    if interface_type == "vpc_endpoint" or
      (interface_type == "interface" and description.start_with?("VPC Endpoint Interface "))
      endpoint_id = description.split(" ")[3]
      symlinks << "/network_interface/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/by-vpce/#{endpoint_id}"
    elsif (interface_type == "network_load_balancer" or
           (interface_type == "interface" and requester_id == "amazon-elb")) and
         description.start_with?("ELB ")

      lb_name = description[4..-1]

      if interface_type == "network_load_balancer" or description.start_with?("ELB app/")
        lb_name = lb_name.split("/")[0..-2].join("/")
      end

      lb = "/load_balancer/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/#{lb_name}"

      relations << {
        metadata: {
          type: "/relation/v1/is_part_of",
          updated_at: updated_at,
        },
        properties: {
          a: path,
          b: lb,
        },
      }
    elsif interface_type == "interface" and requester_id == "amazon-elasticsearch" and description.start_with?("ES ")
      es_name = description.split(" ")[1]

      relations << {
        metadata: {
          type: "/relation/v1/is_part_of",
          updated_at: updated_at,
        },
        properties: {
          a: path,
          b: "/computer/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/elasticsearch/#{es_name}",
        },
      }
    elsif interface_type == "interface" and
         resource["configuration"].has_key?("attachment") and
         resource["configuration"]["attachment"].has_key?("instanceId")
      instance_id = resource["configuration"]["attachment"]["instanceId"]

      relations << {
        metadata: {
          type: "/relation/v1/is_part_of",
          updated_at: updated_at,
        },
        properties: {
          a: path,
          b: "/computer/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/#{instance_id}",
        },
      }
    elsif (interface_type == "interface" and description.start_with?("AWS Lambda VPC ENI")) or interface_type == "lambda"
      name = requester_id.split(":")[1]

      relations << {
        metadata: {
          type: "/relation/v1/is_part_of",
          updated_at: updated_at,
        },
        properties: {
          a: path,
          b: "/computer/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/lambda/#{name}",
        },
      }
    elsif interface_type == "interface" and description.start_with?("aws-K8S-")
      instance_id = requester_id.split(":")[1]

      relations << {
        metadata: {
          type: "/relation/v1/is_part_of",
          updated_at: updated_at,
        },
        properties: {
          a: path,
          b: "/computer/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/#{instance_id}",
        },
      }
    elsif interface_type == "interface" and requester_id == "amazon-elasticache" and description.start_with?("ElastiCache")
      ec_name = description.split(/[ +]/)[1]

      relations << {
        metadata: {
          type: "/relation/v1/is_part_of",
          updated_at: updated_at,
        },
        properties: {
          a: path,
          b: "/computer/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/elasticache/#{ec_name}",
        },
      }
    elsif interface_type == "interface" and
         requester_id == "AROAR7QNRWZMXEDBBSGGK:AmazonEKS" and
         description.start_with?("Amazon EKS ")
      eks_name = description[11..-1]

      relations << {
        metadata: {
          type: "/relation/v1/is_part_of",
          updated_at: updated_at,
        },
        properties: {
          a: path,
          b: "/computer/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/eks/#{eks_name}",
        },
      }
    elsif interface_type == "nat_gateway" and description.start_with?("Interface for NAT Gateway ")
      nat_id = description.split(" ")[4]

      relations << {
        metadata: {
          type: "/relation/v1/is_part_of",
          updated_at: updated_at,
        },
        properties: {
          a: path,
          b: "/computer/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/nat/#{nat_id}",
        },
      }
    elsif interface_type == "interface" and description.start_with?("EFS mount target for ")
      fs_id = description.split(" ")[4]

      relations << {
        metadata: {
          type: "/relation/v1/is_part_of",
          updated_at: updated_at,
        },
        properties: {
          a: path,
          b: "/computer/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/efs/#{fs_id}",
        },
      }
    elsif interface_type == "interface" and (requester_id == "amazon-rds" or requester_id == "062191066759" or requester_id == "082811663747")
    # ignore RDS, relations built in rds_instance_entity
    elsif interface_type == "interface" and requester_id == "493711992759"
    # ignore DMS
    elsif interface_type == "interface" and requester_id == "953619373526"
    # ignore Directory Services
    elsif resource["configuration"]["status"] == "available"
    # ignore unbound ENIs
    else
      raise "Unknown ENI: #{JSON.pretty_generate(resource)}"
    end
  end

  alias_entity(
    {
      metadata: {
        type: "/entity/v1/network_interface",
        updated_at: updated_at,
      },
      properties: {
        provider: "aws",
      }
    },
    id: path,
    aliases: symlinks,
  ) + add_ids_to(relations, base: path)
end

def nat_entity(resource)
  path = "/computer/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/nat/#{resource["resourceId"]}"
  updated_at = DateTime.iso8601(resource["configurationItemCaptureTime"]).rfc3339

  relations = tag_relations(path, updated_at, resource)

  [
    {
      metadata: {
        id: path,
        type: "/entity/v1/computer",
        updated_at: updated_at,
      },
      properties: {
        provider: "aws",
        image: "aws-nat",
      }
    },
  ] + add_ids_to(relations, base: path)
end

def lb_entity(resource)
  name_prefix = ""

  if resource["configuration"].has_key?("type")
    case resource["configuration"]["type"]
    when "network"
      name_prefix = "net/"
    when "application"
      name_prefix = "app/"
    else
      raise "unknown lb type"
    end
  end

  name = name_prefix + resource["resourceName"]
  path = "/load_balancer/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/#{name}"
  updated_at = DateTime.iso8601(resource["configurationItemCaptureTime"]).rfc3339

  relations = tag_relations(path, updated_at, resource)

  [
    {
      metadata: {
        id: path,
        type: "/entity/v1/load_balancer",
        updated_at: updated_at,
      },
      properties: {
        provider: "aws",
        scheme: resource["configuration"]["scheme"],
      }
    },
  ] + add_ids_to(relations, base: path)
end

def rds_instance_entity(resource)
  path = "/computer/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/rds/#{resource["resourceName"]}"
  updated_at = DateTime.iso8601(resource["configurationItemCaptureTime"]).rfc3339

  endpoint_dn = resource["configuration"]["endpoint"]["address"]
  endpoint_addresses = Resolv.getaddresses(endpoint_dn)

  net_relations = endpoint_addresses.map { |address|
    {
      metadata: {
        type: "/relation/v1/is_part_of",
        updated_at: updated_at,
      },
      properties: {
        a: "/network_interface/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/by-ip/#{address}",
        b: path,
      },
    }
  }

  relations = tag_relations(path, updated_at, resource) + net_relations

  [
    {
      metadata: {
        id: path,
        type: "/entity/v1/computer",
        updated_at: updated_at,
      },
      properties: {
        provider: "aws",
        image: "aws-rds",
      }
    },
  ] + add_ids_to(relations, base: path)
end

def lambda_entity(resource)
  path = "/computer/aws/#{resource["accountId"]}/#{resource["awsRegion"]}/lambda/#{resource["resourceId"]}"
  updated_at = DateTime.iso8601(resource["configurationItemCaptureTime"]).rfc3339

  [
    {
      metadata: {
        id: path,
        type: "/entity/v1/computer",
        updated_at: updated_at,
      },
      properties: {
        provider: "aws",
        image: "aws-lambda",
      }
    },
  ] + add_ids_to(tag_relations(path, updated_at, resource), base: path)
end


EFS_BANNED_REGIONS = [ "eu-north-1", "sa-east-1" ]


module Ontology

  module Source

    class AWS

      def sync
        (sync_elasticsearch + sync_elasticache + sync_efs + sync_config).flatten
      end

      def account_id
        @account_id ||= begin
                          client = Aws::STS::Client.new
                          client.get_caller_identity.account
                        end
      end

      def regions
        @regions ||= begin
                       client = Aws::EC2::Client.new
                       client.describe_regions.regions.map { |r| r.region_name }
                     end
      end

      def sync_elasticsearch
        updated_at = DateTime.now.rfc3339

        regions.map { |region|
          client = Aws::ElasticsearchService::Client.new(region: region)
          client.list_domain_names.domain_names.map { |dn|
            domain_name = dn[:domain_name]

            path = "/computer/aws/#{account_id}/#{region}/elasticsearch/#{domain_name}"

            [
              {
                metadata: {
                  id: path,
                  type: "/entity/v1/computer",
                  updated_at: updated_at,
                },
                properties: {
                  provider: "aws",
                  image: "aws-elasticsearch",
                }
              },
            ]
          }
        }.flatten
      end

      def sync_elasticache
        updated_at = DateTime.now.rfc3339

        regions.map { |region|
          client = Aws::ElastiCache::Client.new(region: region)
          client.describe_cache_clusters.cache_clusters.map { |cluster|
            path = "/computer/aws/#{account_id}/#{region}/elasticache/#{cluster.cache_cluster_id}"

            [
              {
                metadata: {
                  id: path,
                  type: "/entity/v1/computer",
                  updated_at: updated_at,
                },
                properties: {
                  provider: "aws",
                  image: "aws-elasticache",
                }
              },
            ]
          }
        }.flatten
      end

      def sync_efs
        updated_at = DateTime.now.rfc3339

        regions.reject { |r| EFS_BANNED_REGIONS.include? r }.map { |region|
          client = Aws::EFS::Client.new(region: region)
          client.describe_file_systems.file_systems.map { |fs|
            path = "/computer/aws/#{account_id}/#{region}/efs/#{fs.file_system_id}"

            [
              {
                metadata: {
                  id: path,
                  type: "/entity/v1/computer",
                  updated_at: updated_at,
                },
                properties: {
                  provider: "aws",
                  image: "aws-efs",
                }
              },
            ]
          }
        }.flatten
      end

      def sync_config
        client = Aws::ConfigService::Client.new
        type_conversion = {
          "AWS::EC2::EIP" => method(:eip_entity),
          "AWS::EC2::Instance" => method(:instance_entity),
          "AWS::EC2::NatGateway" => method(:nat_entity),
          "AWS::EC2::NetworkInterface" => method(:eni_entity),
          "AWS::ElasticLoadBalancing::LoadBalancer" => method(:lb_entity),
          "AWS::ElasticLoadBalancingV2::LoadBalancer" => method(:lb_entity),
          "AWS::RDS::DBInstance" => method(:rds_instance_entity),
          "AWS::Lambda::Function" => method(:lambda_entity),
        }

        next_token = nil
        all_resources = []
        begin
          resources, next_token = find_resources(client, type_conversion.keys, next_token)
          all_resources = all_resources + resources
        end while next_token != "" && next_token != nil

        all_resources.map { |resource|
          type_conversion[resource["resourceType"]].(resource)
        }
      end

      def find_resources(client, types=[], next_token=nil)
        query = "SELECT *, configuration, tags"

        if types.count > 0
          filter = types.map { |type| "resourceType = '#{type}'" }.join(' OR ')
          query += " WHERE " + filter
        end

        resp = client.select_resource_config(
          expression: query,
          next_token: next_token,
          limit: 100,
        )
        return resp[:results].map { |r| JSON.parse(r) }, resp[:next_token]
      end

    end

  end

end
