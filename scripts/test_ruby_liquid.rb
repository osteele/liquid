#!/usr/bin/env ruby
# Test script to verify Shopify Ruby Liquid implementation behavior
# for loop modifiers: reversed, limit, and offset

require 'liquid'

array = [1, 2, 3, 4, 5]

tests = [
  {
    name: "reversed only",
    template: "{% for item in array reversed %}{{ item }}{% endfor %}",
    notes: "Should reverse the entire array"
  },
  {
    name: "limit only",
    template: "{% for item in array limit:2 %}{{ item }}{% endfor %}",
    notes: "Should take first 2 items"
  },
  {
    name: "offset only",
    template: "{% for item in array offset:2 %}{{ item }}{% endfor %}",
    notes: "Should skip first 2 items"
  },
  {
    name: "reversed then limit (syntax order)",
    template: "{% for item in array reversed limit:2 %}{{ item }}{% endfor %}",
    notes: "According to PR#456, should give [5,4] - reverse first, then take first 2"
  },
  {
    name: "limit then reversed (syntax order)",
    template: "{% for item in array limit:2 reversed %}{{ item }}{% endfor %}",
    notes: "Does syntax order matter? Or is application order fixed?"
  },
  {
    name: "limit then offset (syntax order)",
    template: "{% for item in array limit:2 offset:1 %}{{ item }}{% endfor %}",
    notes: "Should offset first, then limit: skip 1, take 2 = [2,3]"
  },
  {
    name: "offset then limit (syntax order)",
    template: "{% for item in array offset:1 limit:2 %}{{ item }}{% endfor %}",
    notes: "Should offset first, then limit: skip 1, take 2 = [2,3]"
  },
  {
    name: "reversed limit offset (syntax order 1)",
    template: "{% for item in array reversed limit:2 offset:1 %}{{ item }}{% endfor %}",
    notes: "How do all three combine?"
  },
  {
    name: "reversed offset limit (syntax order 2)",
    template: "{% for item in array reversed offset:1 limit:2 %}{{ item }}{% endfor %}",
    notes: "Does changing syntax order change result?"
  },
  {
    name: "limit offset reversed (syntax order 3)",
    template: "{% for item in array limit:2 offset:1 reversed %}{{ item }}{% endfor %}",
    notes: "Does reversed at end change things?"
  },
  {
    name: "offset limit reversed (syntax order 4)",
    template: "{% for item in array offset:1 limit:2 reversed %}{{ item }}{% endfor %}",
    notes: "Another permutation"
  },
  {
    name: "offset beyond length",
    template: "{% for item in array offset:10 %}{{ item }}{% endfor %}",
    notes: "Should produce empty result"
  },
  {
    name: "limit 0",
    template: "{% for item in array limit:0 %}{{ item }}{% endfor %}",
    notes: "Should produce empty result"
  },
  {
    name: "reversed with offset beyond length",
    template: "{% for item in array reversed offset:10 %}{{ item }}{% endfor %}",
    notes: "Should produce empty result"
  }
]

puts "=" * 80
puts "Testing Shopify Ruby Liquid Implementation"
puts "Liquid version: #{Liquid::VERSION}"
puts "=" * 80
puts

tests.each_with_index do |test, i|
  template = Liquid::Template.parse(test[:template])
  result = template.render('array' => array)

  puts "#{i+1}. #{test[:name]}"
  puts "   Template: #{test[:template]}"
  puts "   Result:   '#{result}'"
  puts "   Notes:    #{test[:notes]}"
  puts
end
