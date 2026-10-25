-- Managed by HyprProfile.
-- Interactive baseline with common desktop components.

hl.monitor({
    output = "",
    mode = "preferred",
    position = "auto",
    scale = "auto",
})

local terminal = "kitty"
local menu = "wofi --show drun"
local mainMod = "SUPER"

hl.config({
    input = {
        kb_layout = "br",
        kb_model = "abnt2",
        follow_mouse = 1,
    },
    general = {
        layout = "dwindle",
        border_size = 2,
        gaps_in = 5,
        gaps_out = 10,
    },
})

hl.on("hyprland.start", function()
    hl.exec_cmd("sh -lc 'command -v waybar >/dev/null && exec waybar'")
    hl.exec_cmd("sh -lc 'command -v mako >/dev/null && exec mako'")
end)

hl.bind(mainMod .. " + RETURN", hl.dsp.exec_cmd(terminal))
hl.bind(mainMod .. " + D", hl.dsp.exec_cmd(menu))
hl.bind(mainMod .. " + Q", hl.dsp.window.close())
hl.bind(mainMod .. " + SHIFT + M", hl.dsp.exec_cmd("hyprctl dispatch 'hl.dsp.exit()'"))

hl.bind(mainMod .. " + Left", hl.dsp.focus({ direction = "left" }))
hl.bind(mainMod .. " + Right", hl.dsp.focus({ direction = "right" }))
hl.bind(mainMod .. " + Up", hl.dsp.focus({ direction = "up" }))
hl.bind(mainMod .. " + Down", hl.dsp.focus({ direction = "down" }))

for i = 1, 10 do
    local key = i % 10
    hl.bind(mainMod .. " + " .. key, hl.dsp.focus({ workspace = i }))
    hl.bind(mainMod .. " + SHIFT + " .. key, hl.dsp.window.move({ workspace = i }))
end

hl.bind(mainMod .. " + mouse:272", hl.dsp.window.drag(), { mouse = true })
hl.bind(mainMod .. " + mouse:273", hl.dsp.window.resize(), { mouse = true })
