
require 'deep_merge'
require 'digest'
require 'json'
require 'json_schemer'
require 'parallel'
require 'set'
require 'yaml'

class Instance
  def initialize(h)
    @h = h.clone

    if not @h.has_key?("properties")
      @h["properties"] = {}
    end
  end

  def valid?
    not (@h.has_key?("metadata") and
         @h["metadata"].has_key?("id") and
         @h["metadata"].has_key?("type"))
  end

  def id
    @h["metadata"]["id"]
  end

  def name
    if @h["metadata"].has_key?("name")
      @h["metadata"]["name"]
    else
      id.split("/")[-1]
    end
  end

  def type
    @h["metadata"]["type"]
  end

  def properties
    @h["properties"]
  end

  def [](k)
    @h["properties"][k]
  end

  def empty?
    @h["properties"].empty?
  end

  def to_s
    "#{id}[#{type}]: #{properties}"
  end

  def to_str
    to_s
  end
end

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

  attr_reader :relations, :entities, :types

  def initialize
    @add_mutex = Mutex.new
    @relations = []
    @entities = []
    @entities_by_id = {}
    @entities_by_type_id = Hash.new { |h, k| h[k] = [] }
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
    instance = Instance.new(thing)

    raise "Invalid thing: #{thing}" if validate and not valid?(thing)

    @add_mutex.synchronize {
      if instance.type.start_with?("/entity")
        $stderr.puts "Overwriting id #{id}" if @entities_by_id.has_key?(instance.id)

        @entities << instance
        @entities_by_id[instance.id] = instance
        @entities_by_type_id[instance.type] << instance
      elsif instance.type.start_with?("/relation")
        @relations << instance
        @relations_by_id[instance.id] = instance
        @relations_by_entity_id[instance["a"]] << instance
        @relations_by_entity_id[instance["b"]] << instance
      elsif instance.type.start_with?("/type")
        @types << instance
        @types_by_id[instance.id] = instance
      elsif instance.type.start_with?("/link")
        @entities_by_id[id] = instance
      else
        $stderr.puts "Unknown type: #{type}"
      end
    }

    instance
  end

  def valid?(instance, ignore_pointers: false)
    instance = Instance.new(instance) if instance.is_a?(Hash)

    return false if instance.valid?

    return true if instance.id == "/type"

    type = @types_by_id[instance.type]

    return false if not (type and valid?(type))

    type_hierarchy = [type]
    curr_type = type

    while parent_id = curr_type["parent"] and parent = @types_by_id[parent_id]
      type_hierarchy << parent
      curr_type = parent
    end

    merged_spec = type_hierarchy.reverse
                    .map { |t| t["spec"] }
                    .reduce({}, &:deep_merge)

    return true if merged_spec.empty? and not instance.empty?

    schema = {
      "type" => "object",
      "properties" => merged_spec,
    }

    keywords = {}

    keywords = {
      "pointer_to" => ->(data, schema) {
        kind_of?(@entities_by_id[data], schema["pointer_to"])
      },
    }

    schemer = JSONSchemer.schema(
      schema,
      keywords: keywords,
    )

    return schemer.valid?(instance.properties)
  end

  def kind_of?(instance, type_id)
    return false if not instance or not type_id or not @types_by_id.key?(type_id)

    thing_type_id = instance.type

    begin
      return true if thing_type_id = type_id

      thing_type = @types_by_id[thing_type_id]
      thing_type_id = thing_type["properties"]["parent"]
    end while thing_type_id

    return false
  end

  def entities_by_type(type_id)
    @entities_by_type_id[type_id]
  end

  def all_relations_valid?
    @relations.all? { |rel| valid?(rel) }
  end

  def instance_or_id_to_id(instance)
    if instance.is_a?(String)
      id = instance
    elsif instance.is_a?(Instance)
      id = instance.id
    else
      raise "Unknown thing to id: #{instance}"
    end

    id
  end

  def relations_for(instance)
    @relations_by_entity_id[instance_or_id_to_id(instance)]
  end

  def all_relations_for(instance)
    seen = Set.new
    to_traverse = @relations_by_entity_id[instance_or_id_to_id(instance)]

    begin
      to_traverse.each { |rel| seen.add(rel) }

      to_traverse = to_traverse.map { |rel|
        a = rel["a"]
        b = rel["b"]

        a_rels = @relations_by_entity_id[a]
        b_rels = @relations_by_entity_id[b]

        (a_rels + b_rels).compact.reject { |new_rel| seen.include?(new_rel) }
      }.flatten
    end while to_traverse.count > 0

    seen.to_a
  end

  def resolve(instance)
    e = @entities_by_id[instance_or_id_to_id(instance)]
    if e and e.type.start_with? "/link"
      resolve(idx, e["link"])
    elsif e and e.type.start_with? "/entity"
      e
    else
      nil
    end
  end

  def entity_by_id(id)
    @entitiy_by_id[id]
  end

end
