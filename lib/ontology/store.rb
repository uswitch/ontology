
require 'deep_merge'
require 'json'
require 'json_schemer'
require 'set'

require_relative './instance.rb'

module Ontology

  class Store

    attr_reader :relations, :entities, :types

    def initialize
      @add_mutex = Mutex.new
      @relations = []
      @entities = []
      @entities_by_id = {}
      @entities_by_type_id = Hash.new { |h, k| h[k] = [] }
      @relations_by_id = {}
      @relations_by_entity_id = Hash.new { |h, k| h[k] = [] }

      base_type = Instance.new({
                                 metadata: { id: "/type", type: "/type" },
                                 properties: { }
                               })

      @types = [base_type]
      @types_by_id = { "/type" => base_type }
    end

    def add!(thing, validate: false)
      instance = Instance.new(thing)

      raise "Invalid thing: #{thing}" if validate and not valid?(thing)

      @add_mutex.synchronize {
        if instance.type.start_with?("/entity")
          $stderr.puts "Overwriting id #{instance.id}: #{instance.to_h}" if @entities_by_id.has_key?(instance.id)

          @entities << instance
          @entities_by_id[instance.id] = instance
          @entities_by_type_id[instance.type] << instance
        elsif instance.type.start_with?("/relation")
          @relations << instance
          @relations_by_id[instance.id] = instance
          @relations_by_entity_id[instance[:a]] << instance
          @relations_by_entity_id[instance[:b]] << instance
        elsif instance.type.start_with?("/type")
          @types << instance
          @types_by_id[instance.id] = instance
        elsif instance.type.start_with?("/link")
          @entities_by_id[instance.id] = instance
        else
          $stderr.puts "Unknown type: #{type}"
        end
      }

      instance
    end

    def valid?(instance)
      return ! validate(instance).any?
    end

    def validate(instance)
      instance = Instance.new(instance) if instance.is_a?(Hash)

      return ["Not a proper instance"] if instance.valid?

      return [] if instance.id == "/type"

      type = @types_by_id[instance.type]

      return ["No associated, valid type"] if not (type and valid?(type))

      type_hierarchy = [type]
      curr_type = type

      while parent_id = curr_type[:parent] and parent = @types_by_id[parent_id]
        type_hierarchy << parent
        curr_type = parent
      end

      merged_spec = type_hierarchy.reverse
                      .map { |t| t[:spec] }
                      .reduce({}, &:deep_merge)

      return [] if merged_spec.empty? and not instance.empty?

      schema = {
        "type" => "object",
        "properties" => merged_spec,
      }

      keywords = {}

      keywords = {
        :pointer_to => ->(data, schema) {
          type_of?(@entities_by_id[data], schema[:pointer_to])
        },
      }

      schemer = JSONSchemer.schema(
        schema,
        keywords: keywords,
      )

      return schemer.validate(instance.properties)
    end

    def type_of(instance)
      @types_by_id[instance.type]
    end

    def type_of?(instance, type_id)
      return false if not instance or not type_id or not @types_by_id.key?(type_id)

      thing_type_id = instance.type

      begin
        return true if thing_type_id = type_id

        thing_type = @types_by_id[thing_type_id]
        thing_type_id = thing_type[:properties][:parent]
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
      if instance == nil
        id = nil
      elsif instance.is_a?(String)
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
          a = rel[:a]
          b = rel[:b]

          a_rels = @relations_by_entity_id[a]
          b_rels = @relations_by_entity_id[b]

          (a_rels + b_rels).compact.reject { |new_rel| seen.include?(new_rel) }
        }.flatten
      end while to_traverse.count > 0

      seen.to_a
    end

    def type_spec(type_id)
      type = @types_by_id[type_id]
      type_hierarchy = [type]
      curr_type = type

      while parent_id = curr_type[:parent] and parent = @types_by_id[parent_id]
        type_hierarchy << parent
        curr_type = parent
      end

      type_hierarchy.reverse
        .map { |t| t[:spec] }
        .reduce({}, &:deep_merge)
    end

    def resolve(relation)
      rel_spec = type_spec(relation.type)

      a_id = relation[:a]
      a_entity = by_id(a_id)

      if a_entity == nil
        a_entity = Instance.new(
          {
            metadata: {
              type: rel_spec[:a][:pointer_to],
              id: a_id,
            },
            properties: {},
          }
        )
      end

      b_id = relation[:b]
      b_entity = by_id(b_id)

      if b_entity == nil
        $stderr.puts "No pointer_to in: #{rel_spec}" if not rel_spec[:b].has_key?(:pointer_to)
        b_entity = Instance.new(
          {
            metadata: {
              type: rel_spec[:b][:pointer_to],
              id: b_id,
            },
            properties: {},
          }
        )
      end

      return a_entity, b_entity
    end

    def by_id(id)
      @entities_by_id[id]
    end

  end

end
