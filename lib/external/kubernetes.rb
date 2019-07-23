require 'k8s-client'
require 'uri'

require_relative './docker.rb'

def owner_relations(id, server, metadata)
  if metadata.ownerReferences and metadata.ownerReferences.count > 0
    metadata.ownerReferences.map { |ref|
      {
        metadata: {
          type: "/relation/v1/is_part_of",
        },
        properties: {
          a: id,
          b: "/workload/kubernetes/#{server}/#{metadata.namespace}/#{ref.kind.downcase}/#{ref.name}",
        },
      }
    }
  else
    []

  end
end

def container_relations(id, containers)
  containers.map { |container|
    parsed = parse_image_reference(container.image)

    {
      metadata: {
        type: "/relation/v1/supervises",
      },
      properties: {
        a: id,
        b: "/container/#{parsed[:repository]}/#{parsed[:digest] or parsed[:tag]}"
      }
    }
  }
end

class Kubernetes

  def sync
    client = K8s::Client.autoconfig

    server = URI(client.transport.server).host

    nodes = client.api('v1').resource('nodes').list.map do |node|
      id = "/computer/kubernetes/#{server}/#{node.metadata.name}"

      providerId = node.spec.providerID.split("/")
      instanceId = providerId[-1]
      region = providerId[-2][0..-2]

      {
        path: id,
        entity: {
          metadata: {
            type: "/entity/v1/computer",
          },
          properties: {
          },
        },
        relations: [
          {
            metadata: {
              type: "/relation/v1/is_the_same_as",
            },
            properties: {
              a: id,
              b: "/computer/aws/136393635417/#{region}/#{instanceId}",
            },
          },
        ] + (
          labels_to_relations(id, node.metadata.labels.to_h)
        )
      }
    end

    cronjobs = client.api('batch/v1beta1').resource('cronjobs').list.map do |cronjob|
      id = "/workload/kubernetes/#{server}/#{cronjob.metadata.namespace}/cronjobs/#{cronjob.metadata.name}"

      {
        path: id,
        entity: {
          metadata: {
            type: "/entity/v1/workload",
          },
          properties: {
          },
        },
        relations: (
          owner_relations(id, server, cronjob.metadata) +
          labels_to_relations(id, cronjob.metadata.labels.to_h)
        )
      }
    end

    jobs = client.api('batch/v1').resource('jobs').list.map do |job|
      id = "/workload/kubernetes/#{server}/#{job.metadata.namespace}/jobs/#{job.metadata.name}"

      {
        path: id,
        entity: {
          metadata: {
            type: "/entity/v1/workload",
          },
          properties: {
          },
        },
        relations: (
          owner_relations(id, server, job.metadata) +
          labels_to_relations(id, job.metadata.labels.to_h)
        )
      }
    end

    deployments = client.api('apps/v1').resource('deployments').list.map do |deployment|
      id = "/workload/kubernetes/#{server}/#{deployment.metadata.namespace}/deployment/#{deployment.metadata.name}"

      {
        path: id,
        entity: {
          metadata: {
            type: "/entity/v1/workload",
          },
          properties: {
          },
        },
        relations: (
          owner_relations(id, server, deployment.metadata) +
          labels_to_relations(id, deployment.metadata.labels.to_h)
        )
      }
    end

    replica_sets = client.api('apps/v1').resource('replicasets').list.map do |replica_set|
      id = "/workload/kubernetes/#{server}/#{replica_set.metadata.namespace}/replicaset/#{replica_set.metadata.name}"

      {
        path: id,
        entity: {
          metadata: {
            type: "/entity/v1/workload",
          },
          properties: {
          },
        },
        relations: (
          owner_relations(id, server, replica_set.metadata) +
          labels_to_relations(id, replica_set.metadata.labels.to_h)
        )
      }
    end

    daemon_sets = client.api('apps/v1').resource('daemonsets').list.map do |daemon_set|
      id = "/workload/kubernetes/#{server}/#{daemon_set.metadata.namespace}/daemonset/#{daemon_set.metadata.name}"

      {
        path: id,
        entity: {
          metadata: {
            type: "/entity/v1/workload",
          },
          properties: {
          },
        },
        relations: (
          owner_relations(id, server, daemon_set.metadata) +
          labels_to_relations(id, daemon_set.metadata.labels.to_h)
        )
      }
    end

    pods = client.api('v1').resource('pods').list.map do |pod|
      id = "/workload/kubernetes/#{server}/#{pod.metadata.namespace}/pod/#{pod.metadata.name}"

      {
        path: id,
        entity: {
          metadata: {
            type: "/entity/v1/workload",
          },
          properties: {
          },
        },
        relations: [
          {
            metadata: {
              type: "/relation/v1/is_running_on",
            },
            properties: {
              a: id,
              b: "/computer/kubernetes/#{server}/#{pod.spec.nodeName}",
            },
          },
        ] + (
          container_relations(id, pod.spec.containers) +
          owner_relations(id, server, pod.metadata) +
          labels_to_relations(id, pod.metadata.labels.to_h)
        )
      }
    end

    nodes + cronjobs + jobs + deployments + replica_sets + daemon_sets + pods

  end

end
