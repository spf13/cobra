-- Copyright 2013-2023 The Cobra Authors
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--      http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

local native_clink = clink
local generator, hinter, reset
local response, requests, last_command = {':4'}, 0
local refreshes=0
local real_popen = io.popen
clink = {
    version_encoded=10090035,
    generator=function() generator={}; return generator end,
    hinter=function() hinter={}; return hinter end,
    onbeginedit=function(fn) reset=fn end,
    reclassifyline=function() refreshes=refreshes+1 end,
    filematches=function() return {'a.ino','b.INO','c.txt','folder\\'} end,
    dirmatches=function(word) return {{match=word..'child\\',type='dir'}} end,
}
io.popen=function(command)
    requests=requests+1
    last_command=command
    local i=0
    return {lines=function() return function() i=i+1; return response[i] end end,
        close=function() return true end}
end
dofile(arg[1])
assert(generator and hinter and reset)
local function state(words, redir)
    local line=table.concat(words,' ')
    local offsets, offset={},1
    for i,w in ipairs(words) do offsets[i]=offset; offset=offset+#w+1 end
    return {
        getcommandwordindex=function() return 1 end,
        getwordcount=function() return #words end,
        getword=function(_,i) return words[i] end,
        getwordinfo=function(_,i) return {offset=offsets[i],length=#words[i],redir=redir==i} end,
        getendword=function() return words[#words] end,
        getline=function() return line end,
        getcursor=function() return #line+1 end,
    }
end
local function complete(s, lines)
    if lines then response=lines; reset() end
    local b={matches={}}
    function b:addmatch(m) self.matches[#self.matches+1]=m end
    function b:addmatches(ms) for _,m in ipairs(ms) do self:addmatch(m) end end
    function b:setsuppressappend() self.no_space=true end
    function b:setnosort() self.no_sort=true end
    local handled=generator:generate(s,b)
    return b,handled
end
local s=state({'cobrafixture','run',''})
local b,h=complete(s,{'alpha\tFirst','beta\tSecond',':4'})
assert(h and #b.matches==2 and b.matches[1].description=='First')
local expected_request=arg[3]=='no-descriptions' and '__completeNoDesc' or '__complete'
assert(last_command:find(' '..expected_request..' ',1,true),'wrong description protocol')
assert(last_command:find('2>nul "cobrafixture.exe" ',1,true)==1,'request must pin the native executable')
local before=requests
complete(s)
assert(requests==before,'cache miss')
reset(); complete(s)
assert(requests==before+1,'cache was not reset')
b,h=complete(s,{'ignored',':1'}); assert(h and #b.matches==0)
b,h=complete(s,{':0'}); assert(not h)
b,h=complete(s,{':4'}); assert(h)
b=complete(s,{'alpha',':34'}); assert(b.no_space and b.no_sort)
b,h=complete(s,{'ino',':8'}); assert(h and #b.matches==3)
assert(b.matches[1]=='a.ino' and b.matches[2]=='b.INO' and b.matches[3]=='folder\\')
b,h=complete(state({'cobrafixture','run',''}),{'parent',':16'})
assert(h and b.matches[1].match=='child\\')
b,h=complete(s,{'unexpected diagnostic'}); assert(not h and #b.matches==0)
b,h=complete(s,{':64'}); assert(not h)
b=complete(s,{'_activeHelp_ Choose a target',':4'})
assert(#b.matches==0 and hinter:gethint(s)=='Choose a target')
assert(refreshes>0,'Active Help did not request a display refresh')
before=refreshes
complete(s)
assert(refreshes==before,'unchanged hints caused a redundant refresh')
assert(hinter:gethint(state({'cobrafixture','other',''}))==nil)
reset(); assert(hinter:gethint(s)==nil)
for _,unsafe in ipairs({'%PATH%','!PATH!','a"b','a\nb','a\rb'}) do
    before=requests
    complete(state({'cobrafixture','run',unsafe}))
    assert(requests==before,'unsafe shell input reached popen')
end
complete(state({'cobrafixture','run','a & b\\'}),{':4'})
assert(last_command:find('"a & b\\\\"',1,true),'lost trailing backslash or shell quoting')
before=requests
complete(state({'different','run',''})); assert(requests==before)
complete(state({'cobrafixture.cmd','run',''})); assert(requests==before)
complete(state({'cobrafixture'})); assert(requests==before)
complete(state({'cobrafixture','run','out'},3)); assert(requests==before)
print('PASS generated Lua protocol, directives, Active Help, quoting, cache and command isolation')

-- Native Clink line parsing and a freshly built Cobra program on Windows.
if arg[2] and arg[2]~='' then
    assert(native_clink and native_clink.parseline)
    io.popen=real_popen
    local executable=arg[2]
    local function parse(line)
        local all=native_clink.parseline(line)
        local original=all[#all].line_state
        local count=original:getwordcount()
        local empty=line:sub(-1)==' '
        return setmetatable({}, {__index=function(_,key)
            if key=='getwordcount' then return function() return count+(empty and 1 or 0) end end
            if key=='getword' then return function(_,i) return i>count and '' or original:getword(i) end end
            if key=='getwordinfo' then return function(_,i)
                if i>count then return {offset=#line+1,length=0} end
                return original:getwordinfo(i)
            end end
            if key=='getendword' then return function() return empty and '' or original:getendword() end end
            return function(_,...) return original[key](original,...) end
        end})
    end
    local function live(tail,expected)
        reset()
        local line='"'..executable..'" '..tail
        local matches,handled=complete(parse(line))
        assert(handled,'not handled: '..tail)
        if expected then
            local found=false
            for _,m in ipairs(matches.matches) do if m.match==expected then found=true end end
            assert(found,'missing '..expected..' for '..tail)
        end
        return matches,parse(line)
    end
    live('', 'run')
    b=live('run ', 'alpha')
    for _,m in ipairs(b.matches) do
        if m.match=='alpha' then
            if arg[3]=='no-descriptions' then assert(m.description==nil)
            else assert(m.description=='First target') end
        end
    end
    live('--config "a & b\\" run ', 'config:a & b\\')
    live('--config=cpu=atmega328,clock=16MHz run ', 'config:cpu=atmega328,clock=16MHz')
    b=live('run nospace','nospace:'); assert(b.no_space)
    b=live('run order','order-z'); assert(b.no_sort and b.matches[1].match=='order-z')
    b=live('run error'); assert(#b.matches==0)
    b,s=live('run help'); assert(#b.matches==0 and hinter:gethint(s)=='Choose a target')
    print('PASS native Clink parser and real Cobra executable integration')
end
io.popen=real_popen
clink=native_clink
