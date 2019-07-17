
require 'digest'
require 'json'
require 'parallel'
require 'set'

class Store

  def self.from_directories(directories, progress: nil, glob: File.join("**", "*.{json}"))
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

          if not parsed["metadata"].has_key?("id")
            parsed["metadata"]["id"] = id
          end
        end

        store.add!(parsed)
      }
    }

    puts

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
  end

  def add!(thing)
    @add_mutex.synchronize {
      id = thing["metadata"]["id"]
      type = thing["metadata"]["type"]

      if type.start_with?("/entities")
        $stderr.puts "Overwriting id #{id}" if @entities_by_id.has_key?(id)

        @entities << thing
        @entities_by_id[id] = thing
      elsif type.start_with?("/relation")
        @relations << thing
        @relations_by_id[thing["metadata"]["id"]] = thing
        @relations_by_entity_id[thing["properties"]["a"]] << thing
        @relations_by_entity_id[thing["properties"]["b"]] << thing
      elsif type.start_with?("/link")
        @entities_by_id[id] = thing
      else
        $stderr.puts "Unknown type: #{type}"
      end
    }
  end

  def all_relations_for(id)
    seen = Set.new
    to_traverse = @relations_by_entity_id[id]

    puts @relations_by_entitiy_id

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
    elsif e and e["metadata"]["type"].start_with? "/entities"
      e
    else
      nil
    end
  end


end
