<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api/client'
import Button from '@/components/ui/button/Button.vue'
import Input from '@/components/ui/input/Input.vue'
import Label from '@/components/ui/label/Label.vue'
import Card from '@/components/ui/card/Card.vue'
import CardHeader from '@/components/ui/card/CardHeader.vue'
import CardTitle from '@/components/ui/card/CardTitle.vue'
import CardDescription from '@/components/ui/card/CardDescription.vue'
import CardContent from '@/components/ui/card/CardContent.vue'

import { ArrowLeft, Key, Link, Unlink, AlertTriangle, Check } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()

const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const passwordError = ref('')
const passwordSuccess = ref('')
const oidcEnabled = ref(false)
const oidcProviderName = ref('Pocket ID')
const isLinking = ref(false)

onMounted(async () => {
  try {
    const status = await api.getOIDCStatus()
    oidcEnabled.value = status.enabled
    oidcProviderName.value = status.provider_name
    await authStore.fetchUser()
  } catch (err) {
    console.error('Failed to initialize:', err)
  }
})

function goBack() {
  router.push('/')
}

async function changePassword() {
  passwordError.value = ''
  passwordSuccess.value = ''

  if (newPassword.value !== confirmPassword.value) {
    passwordError.value = 'Passwords do not match'
    return
  }

  if (newPassword.value.length < 12) {
    passwordError.value = 'Password must be at least 12 characters'
    return
  }

  try {
    await authStore.changePassword(currentPassword.value, newPassword.value)
    passwordSuccess.value = 'Password changed successfully'
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (err: any) {
    passwordError.value = err.message || 'Failed to change password'
  }
}

async function handleOIDCLink() {
  isLinking.value = true
  try {
    const response = await api.initiateOIDCLink()
    window.location.href = response.auth_url
  } catch (err: any) {
    passwordError.value = 'Failed to initiate OIDC linking'
    isLinking.value = false
  }
}

async function handleOIDCUnlink() {
  if (!confirm('Are you sure you want to unlink your OIDC account?')) return

  try {
    await authStore.unlinkOIDC()
    passwordSuccess.value = 'OIDC account unlinked successfully'
  } catch (err: any) {
    passwordError.value = err.message || 'Failed to unlink OIDC'
  }
}

async function togglePasswordLogin() {
  try {
    const newValue = !authStore.user?.password_enabled
    await authStore.togglePasswordLogin(newValue)
    passwordSuccess.value = `Password login ${newValue ? 'enabled' : 'disabled'}`
  } catch (err: any) {
    passwordError.value = err.message || 'Failed to toggle password login'
  }
}

async function removePassword() {
  if (!confirm('Are you sure? You will only be able to login via OIDC.')) return

  try {
    await authStore.removePassword()
    passwordSuccess.value = 'Password removed. You can only login via OIDC now.'
  } catch (err: any) {
    passwordError.value = err.message || 'Failed to remove password'
  }
}
</script>

<template>
  <div class="min-h-screen bg-background">
    <header class="border-b">
      <div class="container mx-auto px-4 py-4 flex items-center gap-4">
        <Button variant="ghost" @click="goBack">
          <ArrowLeft class="w-4 h-4 mr-2" />
          Back
        </Button>
        <h1 class="text-2xl font-bold">Profile</h1>
      </div>
    </header>

    <main class="container mx-auto px-4 py-8">
      <div class="max-w-2xl mx-auto space-y-6">
        <!-- User Info Card -->
        <Card>
          <CardHeader>
            <CardTitle>Account Information</CardTitle>
            <CardDescription>Your account details and status</CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <Label class="text-muted-foreground">Email</Label>
                <p class="font-medium">{{ authStore.user?.email }}</p>
              </div>
              <div>
                <Label class="text-muted-foreground">Role</Label>
                <p class="font-medium capitalize">{{ authStore.user?.role }}</p>
              </div>
              <div>
                <Label class="text-muted-foreground">Status</Label>
                <p 
                  :class="authStore.user?.is_active ? 'text-green-600' : 'text-red-600'"
                  class="font-medium"
                >
                  {{ authStore.user?.is_active ? 'Active' : 'Inactive' }}
                </p>
              </div>
              <div>
                <Label class="text-muted-foreground">Account Created</Label>
                <p class="font-medium">
                  {{ authStore.user?.created_at ? new Date(authStore.user.created_at).toLocaleDateString() : '-' }}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- Authentication Methods Card -->
        <Card>
          <CardHeader>
            <CardTitle>Authentication Methods</CardTitle>
            <CardDescription>Manage how you can login to your account</CardDescription>
          </CardHeader>
          <CardContent class="space-y-6">
            <!-- Password Section -->
            <div class="space-y-4">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <Key class="w-5 h-5 text-primary" />
                  <div>
                    <p class="font-medium">Password Login</p>
                    <p class="text-sm text-muted-foreground">
                      {{ authStore.user?.password_enabled ? 'Enabled' : 'Disabled' }}
                    </p>
                  </div>
                </div>
                
                <div class="flex gap-2">
                  <Button 
                    v-if="authStore.user?.password_enabled"
                    variant="outline" 
                    size="sm"
                    @click="togglePasswordLogin"
                  >
                    Disable
                  </Button>
                  <Button 
                    v-else
                    variant="outline" 
                    size="sm"
                    @click="togglePasswordLogin"
                  >
                    Enable
                  </Button>
                </div>
              </div>

              <!-- Change Password Form -->
              <div v-if="authStore.user?.password_enabled" class="space-y-4 pl-7">
                <div class="space-y-2">
                  <Label>Current Password</Label>
                  <Input v-model="currentPassword" type="password" />
                </div>

                <div class="space-y-2">
                  <Label>New Password</Label>
                  <Input v-model="newPassword" type="password" />
                </div>

                <div class="space-y-2">
                  <Label>Confirm New Password</Label>
                  <Input v-model="confirmPassword" type="password" />
                </div>

                <Button @click="changePassword">Change Password</Button>

                <Button 
                  v-if="authStore.user?.oidc_enabled"
                  variant="destructive" 
                  size="sm"
                  class="ml-2"
                  @click="removePassword"
                >
                  Remove Password
                </Button>
              </div>
            </div>

            <hr />

            <!-- OIDC Section -->
            <div v-if="oidcEnabled" class="space-y-4">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <Link class="w-5 h-5 text-primary" />
                  <div>
                    <p class="font-medium">{{ oidcProviderName }} Login</p>
                    <p class="text-sm text-muted-foreground">
                      {{ authStore.user?.is_linked_to_oidc ? 'Linked' : 'Not linked' }}
                    </p>
                  </div>
                </div>

                <div class="flex gap-2">
                  <Button
                    v-if="!authStore.user?.is_linked_to_oidc"
                    variant="outline"
                    size="sm"
                    :disabled="isLinking"
                    @click="handleOIDCLink"
                  >
                    <Link class="w-4 h-4 mr-2" />
                    Link Account
                  </Button>
                  
                  <Button
                    v-else
                    variant="outline"
                    size="sm"
                    @click="handleOIDCUnlink"
                  >
                    <Unlink class="w-4 h-4 mr-2" />
                    Unlink
                  </Button>
                </div>
              </div>
            </div>

            <!-- Messages -->
            <div v-if="passwordError" class="flex items-center gap-2 text-red-500 text-sm">
              <AlertTriangle class="w-4 h-4" />
              {{ passwordError }}
            </div>

            <div v-if="passwordSuccess" class="flex items-center gap-2 text-green-600 text-sm">
              <Check class="w-4 h-4" />
              {{ passwordSuccess }}
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  </div>
</template>