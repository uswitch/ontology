
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


def read_json_file(path, id:, mtime:)
  parsed = JSON.parse(File.read(path), symbolize_names: true)

  raise "bad file: #{path}" if not parsed.has_key?(:metadata)

  if not parsed[:metadata].has_key?(:id)
    parsed[:metadata][:id] = id
  end

  if not parsed[:metadata].has_key?(:updated_at)
    parsed[:metadata][:updated_at] = mtime
  end

  parsed
end


module Ontology

  module Source

    class Directory

      def initialize(path, num_threads: 10, glob: File.join("**", "*.{json,yaml}"))
        base = File.expand_path(path)

        if not File.directory? base
          raise "#{base} isn't a directory"
        end

        @base = base
        @num_threads = num_threads
        @glob = glob
      end

      def sync
        files = Dir.glob(@glob, base: @base)

        Parallel.map(files, in_threads: @num_threads) { |file|
          path = File.join(@base, file)
          if (dirname = "#{File.dirname(file)}/") == "./"
            dirname = ""
          end

          id = "/#{dirname}#{File.basename(file, ".*")}"

          mtime = File.stat(path).mtime.to_datetime.rfc3339

          if File.symlink?(path)
            link = File.readlink(path)
            if not link.start_with?(@base)
              $stderr.puts "Link doesn't share a common base with entity: #{path} #{link}"
              next
            end

            link_file = link[@base.length..-1]
            link_id = "#{File.dirname(link_file)}/#{File.basename(link_file, ".*")}"

            parsed = read_json_file(link, id: link_id, mtime: mtime)
          elsif File.extname(file) == ".json"
            parsed = read_json_file(path, id: id, mtime: mtime)
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
        }.flatten
      end

    end

  end

end
