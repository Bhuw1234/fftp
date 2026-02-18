import React, { useState } from 'react'
import { 
  User, 
  Bell, 
  Shield, 
  Key, 
  Palette,
  Download,
  Upload,
  Save,
  CheckCircle
} from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'
import toast from 'react-hot-toast'

interface UserSettings {
  name: string
  email: string
  notifications: {
    email: boolean
    push: boolean
    job_completion: boolean
    node_offline: boolean
    credit_low: boolean
  }
  privacy: {
    profile_public: boolean
    show_earnings: boolean
    share_usage: boolean
  }
  preferences: {
    theme: 'light' | 'dark' | 'system'
    language: string
    timezone: string
    date_format: '12h' | '24h'
  }
  api: {
    two_factor_enabled: boolean
    api_key_visible: boolean
  }
}

const Settings: React.FC = () => {
  const { user } = useAuth()
  const [activeTab, setActiveTab] = useState('profile')
  const [loading, setLoading] = useState(false)
  const [settings, setSettings] = useState<UserSettings>({
    name: user?.name || '',
    email: user?.email || '',
    notifications: {
      email: true,
      push: true,
      job_completion: true,
      node_offline: false,
      credit_low: true
    },
    privacy: {
      profile_public: false,
      show_earnings: false,
      share_usage: true
    },
    preferences: {
      theme: 'system',
      language: 'en',
      timezone: 'UTC',
      date_format: '24h'
    },
    api: {
      two_factor_enabled: false,
      api_key_visible: false
    }
  })

  const handleSave = async () => {
    setLoading(true)
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000))
      toast.success('Settings saved successfully')
    } catch (error) {
      toast.error('Failed to save settings')
    } finally {
      setLoading(false)
    }
  }

  const handleExport = () => {
    const dataStr = JSON.stringify(settings, null, 2)
    const dataBlob = new Blob([dataStr], { type: 'application/json' })
    const url = URL.createObjectURL(dataBlob)
    const link = document.createElement('a')
    link.href = url
    link.download = 'deparrow-settings.json'
    link.click()
    URL.revokeObjectURL(url)
    toast.success('Settings exported successfully')
  }

  const handleImport = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (file) {
      const reader = new FileReader()
      reader.onload = (e) => {
        try {
          const imported = JSON.parse(e.target?.result as string)
          setSettings(imported)
          toast.success('Settings imported successfully')
        } catch (error) {
          toast.error('Invalid settings file')
        }
      }
      reader.readAsText(file)
    }
  }

  const tabs = [
    { id: 'profile', name: 'Profile', icon: User },
    { id: 'notifications', name: 'Notifications', icon: Bell },
    { id: 'privacy', name: 'Privacy', icon: Shield },
    { id: 'preferences', name: 'Preferences', icon: Palette },
    { id: 'api', name: 'API & Security', icon: Key },
  ]

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Settings</h1>
        <p className="text-gray-600">
          Manage your account settings and preferences.
        </p>
      </div>

      <div className="bg-white shadow rounded-lg">
        {/* Tab Navigation */}
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8 px-6">
            {tabs.map((tab) => {
              const Icon = tab.icon
              return (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`flex items-center py-4 px-1 border-b-2 font-medium text-sm ${
                    activeTab === tab.id
                      ? 'border-blue-500 text-blue-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }`}
                >
                  <Icon className="mr-2 h-5 w-5" />
                  {tab.name}
                </button>
              )
            })}
          </nav>
        </div>

        {/* Tab Content */}
        <div className="p-6">
          {activeTab === 'profile' && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-medium text-gray-900">Profile Information</h3>
                <p className="text-sm text-gray-500">
                  Update your personal information and account details.
                </p>
              </div>

              <div className="grid grid-cols-1 gap-6">
                <div>
                  <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                    Full Name
                  </label>
                  <input
                    type="text"
                    id="name"
                    value={settings.name}
                    onChange={(e) => setSettings({...settings, name: e.target.value})}
                    className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                  />
                </div>

                <div>
                  <label htmlFor="email" className="block text-sm font-medium text-gray-700">
                    Email Address
                  </label>
                  <input
                    type="email"
                    id="email"
                    value={settings.email}
                    onChange={(e) => setSettings({...settings, email: e.target.value})}
                    className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                  />
                </div>

                <div>
                  <label htmlFor="timezone" className="block text-sm font-medium text-gray-700">
                    Timezone
                  </label>
                  <select
                    id="timezone"
                    value={settings.preferences.timezone}
                    onChange={(e) => setSettings({
                      ...settings, 
                      preferences: {...settings.preferences, timezone: e.target.value}
                    })}
                    className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                  >
                    <option value="UTC">UTC</option>
                    <option value="America/New_York">Eastern Time</option>
                    <option value="America/Chicago">Central Time</option>
                    <option value="America/Denver">Mountain Time</option>
                    <option value="America/Los_Angeles">Pacific Time</option>
                    <option value="Europe/London">London</option>
                    <option value="Europe/Paris">Paris</option>
                    <option value="Asia/Tokyo">Tokyo</option>
                  </select>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'notifications' && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-medium text-gray-900">Notification Preferences</h3>
                <p className="text-sm text-gray-500">
                  Choose how you want to be notified about events.
                </p>
              </div>

              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="text-sm font-medium text-gray-900">Email Notifications</h4>
                    <p className="text-sm text-gray-500">Receive notifications via email</p>
                  </div>
                  <input
                    type="checkbox"
                    checked={settings.notifications.email}
                    onChange={(e) => setSettings({
                      ...settings,
                      notifications: {...settings.notifications, email: e.target.checked}
                    })}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                </div>

                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="text-sm font-medium text-gray-900">Push Notifications</h4>
                    <p className="text-sm text-gray-500">Receive push notifications in browser</p>
                  </div>
                  <input
                    type="checkbox"
                    checked={settings.notifications.push}
                    onChange={(e) => setSettings({
                      ...settings,
                      notifications: {...settings.notifications, push: e.target.checked}
                    })}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                </div>

                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="text-sm font-medium text-gray-900">Job Completion</h4>
                    <p className="text-sm text-gray-500">Notify when jobs complete or fail</p>
                  </div>
                  <input
                    type="checkbox"
                    checked={settings.notifications.job_completion}
                    onChange={(e) => setSettings({
                      ...settings,
                      notifications: {...settings.notifications, job_completion: e.target.checked}
                    })}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                </div>

                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="text-sm font-medium text-gray-900">Node Offline</h4>
                    <p className="text-sm text-gray-500">Notify when nodes go offline</p>
                  </div>
                  <input
                    type="checkbox"
                    checked={settings.notifications.node_offline}
                    onChange={(e) => setSettings({
                      ...settings,
                      notifications: {...settings.notifications, node_offline: e.target.checked}
                    })}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                </div>

                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="text-sm font-medium text-gray-900">Low Credits</h4>
                    <p className="text-sm text-gray-500">Notify when credits are running low</p>
                  </div>
                  <input
                    type="checkbox"
                    checked={settings.notifications.credit_low}
                    onChange={(e) => setSettings({
                      ...settings,
                      notifications: {...settings.notifications, credit_low: e.target.checked}
                    })}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                </div>
              </div>
            </div>
          )}

          {activeTab === 'privacy' && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-medium text-gray-900">Privacy Settings</h3>
                <p className="text-sm text-gray-500">
                  Control how your information is shared and displayed.
                </p>
              </div>

              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="text-sm font-medium text-gray-900">Public Profile</h4>
                    <p className="text-sm text-gray-500">Make your profile visible to other users</p>
                  </div>
                  <input
                    type="checkbox"
                    checked={settings.privacy.profile_public}
                    onChange={(e) => setSettings({
                      ...settings,
                      privacy: {...settings.privacy, profile_public: e.target.checked}
                    })}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                </div>

                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="text-sm font-medium text-gray-900">Show Earnings</h4>
                    <p className="text-sm text-gray-500">Display your earnings in public stats</p>
                  </div>
                  <input
                    type="checkbox"
                    checked={settings.privacy.show_earnings}
                    onChange={(e) => setSettings({
                      ...settings,
                      privacy: {...settings.privacy, show_earnings: e.target.checked}
                    })}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                </div>

                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="text-sm font-medium text-gray-900">Share Usage Data</h4>
                    <p className="text-sm text-gray-500">Help improve DEparrow by sharing anonymous usage data</p>
                  </div>
                  <input
                    type="checkbox"
                    checked={settings.privacy.share_usage}
                    onChange={(e) => setSettings({
                      ...settings,
                      privacy: {...settings.privacy, share_usage: e.target.checked}
                    })}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                </div>
              </div>
            </div>
          )}

          {activeTab === 'preferences' && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-medium text-gray-900">Display Preferences</h3>
                <p className="text-sm text-gray-500">
                  Customize your interface and regional settings.
                </p>
              </div>

              <div className="grid grid-cols-1 gap-6">
                <div>
                  <label htmlFor="theme" className="block text-sm font-medium text-gray-700">
                    Theme
                  </label>
                  <select
                    id="theme"
                    value={settings.preferences.theme}
                    onChange={(e) => setSettings({
                      ...settings,
                      preferences: {...settings.preferences, theme: e.target.value as 'light' | 'dark' | 'system'}
                    })}
                    className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                  >
                    <option value="light">Light</option>
                    <option value="dark">Dark</option>
                    <option value="system">System</option>
                  </select>
                </div>

                <div>
                  <label htmlFor="language" className="block text-sm font-medium text-gray-700">
                    Language
                  </label>
                  <select
                    id="language"
                    value={settings.preferences.language}
                    onChange={(e) => setSettings({
                      ...settings,
                      preferences: {...settings.preferences, language: e.target.value}
                    })}
                    className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                  >
                    <option value="en">English</option>
                    <option value="es">Español</option>
                    <option value="fr">Français</option>
                    <option value="de">Deutsch</option>
                    <option value="zh">中文</option>
                    <option value="ja">日本語</option>
                  </select>
                </div>

                <div>
                  <label htmlFor="date_format" className="block text-sm font-medium text-gray-700">
                    Time Format
                  </label>
                  <select
                    id="date_format"
                    value={settings.preferences.date_format}
                    onChange={(e) => setSettings({
                      ...settings,
                      preferences: {...settings.preferences, date_format: e.target.value as '12h' | '24h'}
                    })}
                    className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                  >
                    <option value="12h">12-hour (AM/PM)</option>
                    <option value="24h">24-hour</option>
                  </select>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'api' && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-medium text-gray-900">API & Security</h3>
                <p className="text-sm text-gray-500">
                  Manage API access and security settings.
                </p>
              </div>

              <div className="space-y-6">
                {/* API Key Section */}
                <div className="bg-gray-50 rounded-lg p-4">
                  <h4 className="text-sm font-medium text-gray-900 mb-2">API Key</h4>
                  <div className="flex items-center space-x-2">
                    <input
                      type={settings.api.api_key_visible ? 'text' : 'password'}
                      value="sk_deparrow_1234567890abcdef..."
                      readOnly
                      className="flex-1 border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm bg-white"
                    />
                    <button
                      onClick={() => setSettings({
                        ...settings,
                        api: {...settings.api, api_key_visible: !settings.api.api_key_visible}
                      })}
                      className="px-3 py-2 border border-gray-300 rounded-md text-sm text-gray-700 hover:bg-gray-50"
                    >
                      {settings.api.api_key_visible ? 'Hide' : 'Show'}
                    </button>
                    <button className="px-3 py-2 bg-blue-600 text-white rounded-md text-sm hover:bg-blue-700">
                      Regenerate
                    </button>
                  </div>
                  <p className="text-xs text-gray-500 mt-1">
                    Keep your API key secret and secure. It provides full access to your account.
                  </p>
                </div>

                {/* Two-Factor Authentication */}
                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="text-sm font-medium text-gray-900">Two-Factor Authentication</h4>
                    <p className="text-sm text-gray-500">
                      Add an extra layer of security to your account
                    </p>
                  </div>
                  <div className="flex items-center space-x-2">
                    {!settings.api.two_factor_enabled ? (
                      <button className="px-3 py-2 bg-yellow-600 text-white rounded-md text-sm hover:bg-yellow-700">
                        Enable 2FA
                      </button>
                    ) : (
                      <>
                        <CheckCircle className="h-5 w-5 text-green-500" />
                        <span className="text-sm text-green-600">Enabled</span>
                        <button className="px-3 py-2 border border-gray-300 rounded-md text-sm text-gray-700 hover:bg-gray-50">
                          Disable
                        </button>
                      </>
                    )}
                  </div>
                </div>

                {/* Session Management */}
                <div>
                  <h4 className="text-sm font-medium text-gray-900 mb-3">Active Sessions</h4>
                  <div className="space-y-2">
                    <div className="flex items-center justify-between p-3 border rounded-lg">
                      <div>
                        <p className="text-sm font-medium text-gray-900">Current Session</p>
                        <p className="text-xs text-gray-500">Chrome on macOS • San Francisco, CA</p>
                      </div>
                      <div className="flex items-center space-x-2">
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800">
                          Active
                        </span>
                      </div>
                    </div>
                    <div className="flex items-center justify-between p-3 border rounded-lg">
                      <div>
                        <p className="text-sm font-medium text-gray-900">Mobile App</p>
                        <p className="text-xs text-gray-500">iOS • 2 hours ago</p>
                      </div>
                      <button className="text-red-600 hover:text-red-800 text-sm">
                        Revoke
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Action Buttons */}
        <div className="border-t border-gray-200 px-6 py-4">
          <div className="flex justify-between">
            <div className="flex space-x-3">
              <button
                onClick={handleExport}
                className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              >
                <Download className="mr-2 h-4 w-4" />
                Export Settings
              </button>
              <label className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 cursor-pointer">
                <Upload className="mr-2 h-4 w-4" />
                Import Settings
                <input
                  type="file"
                  accept=".json"
                  onChange={handleImport}
                  className="hidden"
                />
              </label>
            </div>
            <button
              onClick={handleSave}
              disabled={loading}
              className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
            >
              {loading ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Saving...
                </>
              ) : (
                <>
                  <Save className="mr-2 h-4 w-4" />
                  Save Changes
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Settings