# frozen_string_literal: true
#
# Usage of Observable — the publish/subscribe mixin backed by this library.
# A publisher includes Observable, marks itself changed, and notifies its
# observers, each of which is dispatched through its #update method (in the
# order it was added). Runs under go-embedded-ruby (rbgo); see examples/README.md.

require "observer"

# The publisher mixes in Observable and drives the changed/notify lifecycle.
class Thermostat
  include Observable

  def temperature=(celsius)
    changed                        # mark state changed (default: true)
    notify_observers(celsius)      # dispatch #update to each observer, then reset
  end
end

# An observer just needs to respond to the notification method (default :update).
class Display
  def initialize(name) = @name = name
  def update(celsius) = puts "#{@name}: #{celsius}C"
end

station = Thermostat.new
p station.add_observer(Display.new("kitchen"))  # => :update  (the notify method)
p station.add_observer(Display.new("garage"))   # => :update
p station.count_observers                        # => 2

p station.changed?                               # => false  (nothing pending yet)
station.temperature = 21                         # => kitchen: 21C / garage: 21C
p station.changed?                               # => false  (notify reset the flag)

p station.delete_observers                       # remove every observer -> {}
p station.count_observers                        # => 0
p station.notify_observers(99)                   # => nil   (not changed: a no-op)
