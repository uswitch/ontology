

module Ontology


  class Instance
    def initialize(h)
      @h = h.clone

      if not @h.has_key?(:properties)
        @h[:properties] = {}
      end

      if not @h[:metadata].has_key?(:name)
        @h[:metadata][:name] = id.split("/")[-1]
      end

      if not @h[:metadata].has_key?(:updated_at)
        @h[:metadata][:updated_at] = DateTime.now.rfc3339
      end

      raise "Invalid instance #{h}" if not valid?
    end

    def valid?
      return false if not (@h.has_key?(:metadata) and
                           @h[:metadata].has_key?(:id) and
                           @h[:metadata].has_key?(:type) and
                           @h[:metadata].has_key?(:name) and
                           @h[:metadata].has_key?(:updated_at))

      # and updated_at is in RFC3339 format
      begin
        updated_at = DateTime.rfc3339(@h[:metadata][:updated_at])
      rescue ArgumentError
        return false
      end

      return true
    end

    def id
      @h[:metadata][:id]
    end

    def name
      if @h[:metadata].has_key?(:name)
        @h[:metadata][:name]
      else
        id.split("/")[-1]
      end
    end

    def type
      $stderr.puts @h if not @h.has_key?(:metadata)
      @h[:metadata][:type]
    end

    def updated_at
      @h[:metadata][:updated_at]
    end

    def properties
      @h[:properties]
    end

    def [](k)
      @h[:properties][k]
    end

    def empty?
      @h[:properties].empty?
    end

    def to_h
      @h
    end

    def to_s
      #"#{id}[#{type}]: #{properties}"
      @h.to_s
    end

    def to_str
      to_s
    end
  end

end
