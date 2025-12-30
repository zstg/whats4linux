import { create } from "zustand"
import { GetSettings, SaveSettings } from "../../wailsjs/go/api/Api"

interface AppSettingsStore extends AppSettings {
  loaded: boolean

  loadSettings: () => Promise<void>
  updateSetting: <K extends keyof AppSettings>(key: K, value: AppSettings[K]) => Promise<void>
  toggleTheme: () => Promise<void>
}

export interface AppSettings {
  // Theme
  theme: "light" | "dark"

  // Privacy Settings
  readReceipts: boolean
  blockUnknown: boolean
  disableLinkPreviews: boolean

  // Notifications Settings
  messageNotifications: boolean
  showPreviews: boolean
  showReactionNotifications: boolean
  statusReactions: boolean
  callNotifications: boolean
  incomingCallSounds: boolean
  incomingSounds: boolean
  outgoingSounds: boolean

  // General Settings
  startAtLogin: boolean
  minimizeToTray: boolean
  language: string
  fontSize: string

  // Chats Settings
  spellCheck: boolean
  replaceTextWithEmojis: boolean
  enterIsSend: boolean
}

const defaultSettings: AppSettings = {
  theme: "light",

  readReceipts: true,
  blockUnknown: false,
  disableLinkPreviews: false,

  messageNotifications: true,
  showPreviews: true,
  showReactionNotifications: true,
  statusReactions: true,
  callNotifications: true,
  incomingCallSounds: true,
  incomingSounds: true,
  outgoingSounds: true,

  startAtLogin: false,
  minimizeToTray: true,
  language: "English",
  fontSize: "100% (Default)",

  spellCheck: true,
  replaceTextWithEmojis: true,
  enterIsSend: false,
}

function extractSettings(state: AppSettingsStore): AppSettings {
  const { loaded, ...settings } = state
  return settings
}

export const useAppSettingsStore = create<AppSettingsStore>((set, get) => ({
  ...defaultSettings,
  loaded: false,

  loadSettings: async () => {
    try {
      const saved = await GetSettings()

      const merged = {
        ...defaultSettings,
        ...(saved ?? {}),
      }

      set({
        ...merged,
        loaded: true,
      })
    } catch (err) {
      console.error("Failed to load settings:", err)
      set({ loaded: true })
    }
  },

  updateSetting: async (key, value) => {
    set(state => {
      const next = { ...state, [key]: value }

      SaveSettings(extractSettings(next)).catch(err => {
        console.error("Failed to save setting:", err)
      })

      return next
    })
  },

  toggleTheme: async () => {
    const theme = get().theme === "light" ? "dark" : "light"
    await get().updateSetting("theme", theme)
  },
}))
