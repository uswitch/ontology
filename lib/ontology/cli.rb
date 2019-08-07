
require 'parallel'
require 'yaml'
require 'digest'



module SymbolizeHelper
  extend self

  def symbolize_recursive(hash)
    {}.tap do |h|
      hash.each { |key, value| h[key.to_sym] = transform(value) }
    end
  end

  private

  def transform(thing)
    case thing
    when Hash; symbolize_recursive(thing)
    when Array; thing.map { |v| transform(v) }
    else; thing
    end
  end

  refine Hash do
    def deep_symbolize_keys
      SymbolizeHelper.symbolize_recursive(self)
    end
  end

end

using SymbolizeHelper



module Ontology

  module CLI

    def self.store_from_paths(paths, options)
      directories = []
      files = []

      paths.each { |path|
        if File.directory?(path)
          directories << path
        elsif File.file?(path)
          files << path
        else
          $stderr.puts "Unknown path type: #{path}"
        end
      }

      store = Store.new

      add_directories!(store, directories, options)

      files.each { |file|
        $stderr.puts "Loading #{file}"
        File.readlines(file).each { |thing|
          parsed_thing = JSON.parse(thing, symbolize_names: true)
          $stderr.puts "Problem in #{file} #{parsed_thing}" if not parsed_thing.has_key?(:metadata)
          store.add!(parsed_thing)
        }
      }

      store
    end

  end

end



def add_directories!(store, directories, progress: nil, validate: true, glob: File.join("**", "*.{json,yaml}"))
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

      mtime = File.stat(path).mtime.to_datetime.rfc3339

      if File.symlink?(path)
        link = File.readlink(path)
        if not link.start_with?(base)
          $stderr.puts "Link doesn't share a common base with entity: #{path} #{link}"
          next
        end

        link_file = link[base.length..-1]
        link_id = "#{File.dirname(link_file)}/#{File.basename(link_file, ".*")}"

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
        parsed = JSON.parse(File.read(path), symbolize_names: true)

        raise "bad file: #{path}" if not parsed.has_key?(:metadata)

        if not parsed[:metadata].has_key?(:id)
          parsed[:metadata][:id] = id
        end

        if not parsed[:metadata].has_key?(:updated_at)
          parsed[:metadata][:updated_at] = mtime
        end
      elsif File.extname(file) == ".yaml"
        parsed = []
        File.open( path ) do |yf|
          idx = 0
          YAML.load_stream( yf ) do |ydoc_raw|

            ydoc = ydoc_raw.deep_symbolize_keys

            if not ydoc[:metadata].has_key?(:id)
              suffix = ""
              if idx > 0
                suffix = "/#{idx}"
              end
              ydoc[:metadata][:id] = "#{id}#{suffix}"
            end

            if not ydoc[:metadata].has_key?(:updated_at)
              ydoc[:metadata][:updated_at] = mtime
            end

            parsed << ydoc

            idx += 1
          end
        end
      end

      parsed
    }
  }.flatten

  if progress
    opts = progress.clone
    opts[:title] = "Loading into store"
    opts[:total] = all_things.count

    load_progress = ProgressBar.create(**opts)
  end

  all_things.each { |thing|
    store.add!(thing)
    load_progress.increment
  }

  return store
end
