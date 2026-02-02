--- === JTT ===
---
--- Voice-to-text transcription triggered by holding [ and ] together.

local obj = {}
obj.__index = obj

obj.name = "JTT"
obj.version = "1.0"
obj.author = "Justin May"
obj.license = "MIT"

obj._recording = false
obj._leftBracketDown = false
obj._rightBracketDown = false
obj._tap = nil

function obj:start()
    if not self._recording then
        self._recording = true
        hs.task.new("/usr/bin/env", nil, {"jtt", "start"}):start()
    end
end

function obj:stop()
    if self._recording then
        self._recording = false
        hs.task.new("/usr/bin/env", function(exitCode, stdOut, stdErr)
            if exitCode == 0 then
                hs.timer.doAfter(0.1, function()
                    hs.eventtap.keyStroke({"cmd"}, "v")
                end)
            end
        end, {"jtt", "stop"}):start()
    end
end

function obj:_checkBothKeys()
    if self._leftBracketDown and self._rightBracketDown then
        self:start()
    elseif self._recording and (not self._leftBracketDown or not self._rightBracketDown) then
        self:stop()
    end
end

function obj:init()
    local self = self
    self._tap = hs.eventtap.new({hs.eventtap.event.types.keyDown, hs.eventtap.event.types.keyUp}, function(event)
        local keyCode = event:getKeyCode()
        local eventType = event:getType()
        
        -- [ is keyCode 33, ] is keyCode 30
        if keyCode == 33 then
            if eventType == hs.eventtap.event.types.keyDown then
                self._leftBracketDown = true
            else
                self._leftBracketDown = false
            end
            self:_checkBothKeys()
            if self._leftBracketDown and self._rightBracketDown then
                return true
            end
        elseif keyCode == 30 then
            if eventType == hs.eventtap.event.types.keyDown then
                self._rightBracketDown = true
            else
                self._rightBracketDown = false
            end
            self:_checkBothKeys()
            if self._leftBracketDown and self._rightBracketDown then
                return true
            end
        end
        
        return false
    end)
    
    self._tap:start()
    return self
end

return obj
