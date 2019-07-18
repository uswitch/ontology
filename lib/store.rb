
require 'deep_merge'
require 'digest'
require 'json'
require 'json_schemer'
require 'parallel'
require 'set'

class Store

  def self.from_directories(directories, progress: nil, validate: true, glob: File.join("**", "*.{json}"))
    store = Store.new

    directories.each { |directory|

      base = File.expand_path(directory)

      if not File.directory? base
        raise "#{base} isn't a directory"
      end

      files = Dir.glob(glob, base: base)

      Parallel.each(files, in_threads: 10, progress: progress) { |file|
        path = File.join(base, file)
        id = "/#{File.dirname(file)}/#{File.basename(file, ".*")}"

        if File.symlink?(path)
          link = File.readlink(path)
          if not link.start_with?(base)
            $stderr.puts "Link doesn't share a common base with entity: #{path} #{link}"
            next
          end

          link_file = link[base.length..-1]
          link_id = "/#{File.dirname(link_file)}/#{File.basename(link_file, ".*")}"

          parsed = {
            "metadata" => {
              "type" => "/link",
              "id" => id,
            },
            "properties" => {
              "link" => link_id,
            },
          }
        else
          parsed = JSON.parse(File.read(path))

          raise "bad file: #{path}" if not parsed.has_key?("metadata")

          if not parsed["metadata"].has_key?("id")
            parsed["metadata"]["id"] = id
          end
        end

        # we can't follow pointers until everything is loaded
        store.add!(parsed, validate: validate, ignore_pointers: true)
      }
    }

    invalid_relations = store.relations.reject { |r| store.valid?(r) }.map { |r| r["metadata"]["id"] }

    raise "Invalid relations: #{invalid_relations}" if not invalid_relations.empty?

    store
  end

  attr_reader :relations, :entities

  def initialize
    @add_mutex = Mutex.new
    @relations = []
    @entities = []
    @entities_by_id = {}
    @relations_by_id = {}
    @relations_by_entity_id = Hash.new { |h, k| h[k] = [] }

    base_type = {
      "metadata" => { "id" => "/type", "type" => "/type" },
      "properties" => { }
    }

    @types = [base_type]
    @types_by_id = { "/type" => base_type, "/link" => base_type }
  end

  def add!(thing, validate: true, ignore_pointers: false)
    id = thing["metadata"]["id"]
    type = thing["metadata"]["type"]

    raise "Invalid thing: #{thing}" if validate and not valid?(thing, ignore_pointers: ignore_pointers)

    @add_mutex.synchronize {
      if type.start_with?("/entity")
        $stderr.puts "Overwriting id #{id}" if @entities_by_id.has_key?(id)

        @entities << thing
        @entities_by_id[id] = thing
      elsif type.start_with?("/relation")
        @relations << thing
        @relations_by_id[thing["metadata"]["id"]] = thing
        @relations_by_entity_id[thing["properties"]["a"]] << thing
        @relations_by_entity_id[thing["properties"]["b"]] << thing
      elsif type.start_with?("/type")
        @types << thing
        @types_by_id[thing["metadata"]["id"]] = thing
      elsif type.start_with?("/link")
        @entities_by_id[id] = thing
      else
        $stderr.puts "Unknown type: #{type}"
      end
    }
  end

  def valid?(thing, ignore_pointers: false)
    return false if not (thing.has_key?("metadata") and
                         thing["metadata"].has_key?("id") and
                         thing["metadata"].has_key?("type"))

    return true if thing["metadata"]["id"] == "/type"

    type = @types_by_id[thing["metadata"]["type"]]

    return false if not (type and valid?(type))

    type_hierarchy = [type]
    curr_type = type

    while parent_id = curr_type["properties"]["parent"] and parent = @types_by_id[parent_id]
      type_hierarchy << parent
      curr_type = parent
    end

    merged_spec = type_hierarchy.reverse
                    .map { |t| t["properties"]["spec"] }
                    .reduce({}, &:deep_merge)

    return true if merged_spec.empty? and not thing.has_key?("properties")

    schema = {
      "type" => "object",
      "properties" => merged_spec,
    }

    keywords = {}

    if not ignore_pointers
      keywords = {
        "pointer_to" => ->(data, schema) {
          kind_of?(@entities_by_id[data], schema["pointer_to"])
        },
      }
    end

    schemer = JSONSchemer.schema(
      schema,
      keywords: keywords,
    )

    return schemer.valid?(thing["properties"])
  end

  def kind_of?(thing, type_id)
    return false if not thing or not type_id or not @types_by_id.key?(type_id)

    thing_type_id = thing["metadata"]["type"]

    begin
      return true if thing_type_id = type_id

      thing_type = @types_by_id[thing_type_id]
      thing_type_id = thing_type["properties"]["parent"]
    end while thing_type_id

    return false
  end

  def all_relations_valid?
    @relations.all? { |rel| valid?(rel) }
  end

  def all_relations_for(id)
    seen = Set.new
    to_traverse = @relations_by_entity_id[id]

    begin
      to_traverse.each { |rel| seen.add(rel) }

      to_traverse = to_traverse.map { |rel|
        a = rel["properties"]["a"]
        b = rel["properties"]["b"]

        a_rels = @relations_by_entity_id[a]
        b_rels = @relations_by_entity_id[b]

        (a_rels + b_rels).compact.reject { |new_rel| seen.include?(new_rel) }
      }.flatten
    end while to_traverse.count > 0

    seen.to_a
  end

  def resolve(id)
    e = @entities_by_id[id]
    if e and e["metadata"]["type"].start_with? "/link"
      resolve(idx, e["properties"]["link"])
    elsif e and e["metadata"]["type"].start_with? "/entity"
      e
    else
      nil
    end
  end


end
