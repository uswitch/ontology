
require_relative './source.rb'
require_relative './store.rb'

module Ontology

  module CLI

    def self.store_from_paths(paths, options)
      store = Store.new

      paths.each { |path|
        if File.directory?(path)
          $stderr.puts "Loading directory '#{path}'"
          directory = Ontology::Source::Directory.new(path)
          things = directory.sync
        elsif File.file?(path)
          $stderr.puts "Loading file '#{path}'"
          things = File.readlines(path).map { |thing|
            JSON.parse(thing, symbolize_names: true)
          }
        else
          $stderr.puts "Unknown path type: #{path}"
          next
        end

        $stderr.puts "Adding #{things.count} things to store"
        things.each { |thing| store.add!(thing) }
      }

      store
    end

  end

end
