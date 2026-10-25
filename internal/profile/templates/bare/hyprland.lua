-- Managed by HyprProfile.
-- Small operable baseline without optional desktop components.

hl.monitor({
    output = "",
    mode = "preferred",
    position = "auto",
    scale = "auto",
})

local terminal = "kitty"
local mainMod = "SUPER"

hl.config({
    input = {
        kb_layout = "br",
        kb_model = "abnt2",
        follow_mouse = 1,
    },
    general = {
        layout = "dwindle",
    },
})

hl.bind(mainMod .. " + RETURN", hl.dsp.exec_cmd(terminal))
hl.bind(mainMod .. " + Q", hl.dsp.window.close())
hl.bind(mainMod .. " + SHIFT + M", hl.dsp.exec_cmd("hyprctl dispatch 'hl.dsp.exit()'"))

for i = 1, 10 do
    local key = i % 10
    hl.bind(mainMod .. " + " .. key, hl.dsp.focus({ workspace = i }))
    hl.bind(mainMod .. " + SHIFT + " .. key, hl.dsp.window.move({ workspace = i }))
end

hl.bind(mainMod .. " + mouse:272", hl.dsp.window.drag(), { mouse = true })
hl.bind(mainMod .. " + mouse:273", hl.dsp.window.resize(), { mouse = true })
