local go = require("go")

local ch = channel.make()
local lock = go.lock()

local t = { x = 3 }
for i = 1, count do
    go.go(function(x, v)
          if go.debug then
             lock:lock()
             print(t, x, i, v)
             lock:unlock()
          end
          lock:lock()
          x[tostring(v)] = v
          lock:unlock()
          assert(x.x == t.x, "table field")
          ch:send(i)
          coroutine.yield(v)
          return v
          end, t, i)
end

local sum = 0
local done = count
repeat
   channel.select({"|<-", ch,
                   function(ok, data)
                      sum = sum + data
                      done = done - 1
                   end
   })
until done == 0

ch:close()

if go.debug then
   print("==============")
   for i, line in pairs(t) do
      print(i, line)
   end
   print("==============")
end

for i=1, count do
   k = tostring(i)
   if go.debug then print(i, t[k]) end
   assert(t[k] == i, "table value")
end


return sum
