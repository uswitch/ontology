require 'digest'

TAG_PREFIX="cloud.rvu.ontology"

def clone_hash(hash)
  Marshal.load(Marshal.dump(hash))
end

def alias_entity(entity, id:, aliases:)
  entity_ids = [id] + aliases

  entities = entity_ids.map { |entity_id|
    e = clone_hash(entity)
    e[:metadata][:id] = entity_id
    e
  }

  relations = aliases.map { |alias_id|
    {
      metadata: {
        id: "#{id}/#{Digest::SHA256.hexdigest(alias_id)}",
        type: "/relation/v1/is_the_same_as",
      },
      properties: {
        a: alias_id,
        b: id,
      },
    }
  }

  entities + relations
end

def add_ids_to(things, base:)
  things.each_with_index.map { |thing, i|
    thing[:metadata][:id] = "#{base}/#{i}"
    thing
  }
end

def labels_to_relations(entitiy_id, updated_at, labels)
  labels.map { |tag, val|
    next if not tag.to_s.start_with? TAG_PREFIX

    type = tag[TAG_PREFIX.length..-1].gsub(/\./,"/")

    begin
      vals = JSON.load(val)
      raise "val should be an array: #{val}" if not vals.is_a? Array
    rescue JSON::ParserError
      vals = [val]
    end

    vals.map { |val|
      {
        metadata: {
          type: type,
          updated_at: updated_at,
        },
        properties: {
          a: entitiy_id,
          b: val,
        },
      }
    }
   }.flatten.compact
end
