
require 'deep_merge'
require 'digest'
require 'json'
require 'json_schemer'
require 'parallel'
require 'set'
require 'yaml'

class Store

  def self.from_directories(directories, progress: nil, validate: true, glob: File.join("**", "*.{json,yaml}"))
    store = Store.new

    all_things = directories.map { |directory|

      base = File.expand_path(directory)

      if not File.directory? base
        raise "#{base} isn't a directory"
      end

      files = Dir.glob(glob, base: base)

      if progress
        parse_progress = progress.clone
        parse_progress[:title] = "Parsing files from #{base}"
      end

      Parallel.map(files, in_threads: 10, progress: parse_progress) { |file|
        path = File.join(base, file)
        if (dirname = "#{File.dirname(file)}/") == "./"
          dirname = ""
        end

        id = "/#{dirname}#{File.basename(file, ".*")}"

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
        elsif File.extname(file) == ".json"
          parsed = JSON.parse(File.read(path))

          raise "bad file: #{path}" if not parsed.has_key?("metadata")

          if not parsed["metadata"].has_key?("id")
            parsed["metadata"]["id"] = id
          end
        elsif File.extname(file) == ".yaml"
          parsed = []
          File.open( path ) do |yf|
            idx = 0
            YAML.load_stream( yf ) do |ydoc|
              if not ydoc["metadata"].has_key?("id")
                ydoc["metadata"]["id"] = "#{id}/#{idx}"
              end

              parsed << ydoc

              idx += 1
            end
          end
        end

        parsed
      }
    }.flatten

    all_types = []
    all_entities = []
    all_relations = []

    if progress
      opts = progress.clone
      opts[:title] = "Loading into store"
      opts[:total] = all_things.count * 2

      load_progress = ProgressBar.create(**opts)
    end

    all_things.each { |thing|
      if thing["metadata"]["type"].start_with? "/entity" or
        thing["metadata"]["type"].start_with? "/link"
        all_entities << thing
      elsif thing["metadata"]["type"].start_with? "/relation"
        all_relations << thing
      elsif thing["metadata"]["type"].start_with? "/type"
        all_types << thing
      else
        raise "Unknown type '#{thing["metadata"]["type"]}' for '#{thing["metadata"]["id"]}"
      end

      load_progress.increment
    }

    all_types.each { |thing|
      store.add!(thing)
      load_progress.increment
    }
    all_entities.each { |thing|
      store.add!(thing)
      load_progress.increment
    }
    all_relations.each { |thing|
      if store.valid?(thing)
        store.add!(thing, validate: false)
      else
        $stderr.puts "Dropping invalid relation: #{thing}'"
      end

      load_progress.increment
    }

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

  def add!(thing, validate: true)
    id = thing["metadata"]["id"]
    type = thing["metadata"]["type"]

    raise "Invalid thing: #{thing}" if validate and not valid?(thing)

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
