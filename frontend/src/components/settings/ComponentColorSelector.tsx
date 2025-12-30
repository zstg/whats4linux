import { useEffect, useState } from "react"
import DropDown from "./DropDown"
import ToggleButton from "./ToggleButton"
import { SingleShortcut } from "../../screens/settingscreens/KeyBoardShortCuts"

import { THEME } from "../../theme.config"
import { useAppSettingsStore } from "../../store/useAppSettingsStore"

type ThemeGroup = keyof typeof THEME
type ThemeLabel<G extends ThemeGroup> = keyof (typeof THEME)[G]

const isHex = (v: string) => /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.test(v)

const ComponentColorSelector = () => {
  const { themeColors, updateThemeColor } = useAppSettingsStore()
  const [selectedComponent, setSelectedComponent] = useState<ThemeGroup | null>(null)
  const [draft, setDraft] = useState<typeof themeColors | null>(null)

  useEffect(() => {
    if (selectedComponent) {
      setDraft(structuredClone(themeColors))
    }
  }, [selectedComponent, themeColors])

  const handleDraftChange = <G extends ThemeGroup>(
    group: G,
    label: ThemeLabel<G>,
    value: string,
  ) => {
    setDraft(prev => (prev ? { ...prev, [group]: { ...prev[group], [label]: value } } : prev))
  }

  const saveChanges = async () => {
    if (!selectedComponent || !draft) return
    const group = selectedComponent
    const entries = Object.entries(draft[group]) as Array<[ThemeLabel<typeof group>, string]>

    for (const [label, value] of entries) {
      if (isHex(value)) {
        await updateThemeColor(group, label, value)
      }
    }
  }

  const resetToDefaults = () => {
    if (!selectedComponent || !draft) return
    const group = selectedComponent
    setDraft(prev => (prev ? { ...prev, [group]: { ...THEME[group] } } : prev))
  }

  return (
    <div className="flex flex-row gap-8 border p-4 rounded-md w-full">
      {/* LEFT */}
      <div className="flex flex-col justify-between w-1/3">
        <DropDown
          title=""
          elements={Object.keys(THEME)}
          onToggle={val => setSelectedComponent(val as ThemeGroup)}
          placeholder="Select the Component"
        />

        {selectedComponent && (
          <div className="flex gap-2">
            <button
              className="px-3 py-1 text-sm rounded bg-emerald-500 text-white"
              onClick={saveChanges}
            >
              Save Changes
            </button>

            <button
              className="px-3 py-1 text-sm rounded bg-gray-200 dark:bg-gray-700"
              onClick={resetToDefaults}
            >
              Reset to Defaults
            </button>
          </div>
        )}
      </div>

      {/* MIDDLE */}
      <div className="flex flex-col gap-4 w-1/3">
        {!selectedComponent && (
          <div className="opacity-70 self-center">Select a component to customize 🎨</div>
        )}

        {selectedComponent && draft && (
          <>
            <h2 className="font-semibold text-lg">{selectedComponent} Colors</h2>

            {(
              Object.keys(THEME[selectedComponent]) as Array<ThemeLabel<typeof selectedComponent>>
            ).map(label => {
              const value = draft[selectedComponent][label]

              return (
                <div key={label} className="flex items-center gap-3">
                  <label className="uppercase text-sm opacity-80 w-60">{label}</label>

                  <input
                    type="text"
                    value={value}
                    onChange={e => handleDraftChange(selectedComponent, label, e.target.value)}
                    className="border rounded px-2 py-1 w-32"
                  />

                  <input
                    type="color"
                    value={isHex(value) ? value : "#000000"}
                    onChange={e => handleDraftChange(selectedComponent, label, e.target.value)}
                    className="w-10 h-8 cursor-pointer"
                  />
                </div>
              )
            })}
          </>
        )}
      </div>

      {/* RIGHT */}
      <div className="w-1/3 flex justify-center items-center">
        {selectedComponent === "Keyboard Shortcut" && (
          <SingleShortcut name="Sample" shortcut={["Ctrl", "Shift", "A"]} />
        )}

        {selectedComponent === "Button" && <ToggleButton isEnabled={false} onToggle={() => {}} />}

        {selectedComponent === "DropDown" && (
          <DropDown elements={["A", "B", "C"]} title="DropDown" onToggle={() => {}} />
        )}
      </div>
    </div>
  )
}

export default ComponentColorSelector
