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
import CardFooter from '@/components/ui/card/CardFooter.vue'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const error = ref('')
const isLoading = ref(false)
const oidcEnabled = ref(false)
const oidcProviderName = ref('Pocket ID')

onMounted(async () => {
  try {
    const status = await api.getOIDCStatus()
    oidcEnabled.value = status.enabled
    oidcProviderName.value = status.provider_name
  } catch (err) {
    console.error('Failed to fetch OIDC status:', err)
  }
})

async function handleLogin() {
  error.value = ''
  isLoading.value = true

  try {
    await authStore.login({
      email: email.value,
      password: password.value,
    })
    router.push('/')
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Login failed'
  } finally {
    isLoading.value = false
  }
}

async function handleOIDCLogin() {
  try {
    const response = await api.initiateOIDCLogin()
    window.location.href = response.auth_url
  } catch (err: any) {
    error.value = 'Failed to initiate OIDC login'
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-background p-4">
    <Card class="w-full max-w-md">
      <CardHeader class="space-y-1">
        <CardTitle class="text-2xl font-bold">Welcome back</CardTitle>
        <CardDescription>
          Enter your credentials to access your account
        </CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <form @submit.prevent="handleLogin" class="space-y-4">
          <div class="space-y-2">
            <Label for="email">Email</Label>
            <Input
              id="email"
              v-model="email"
              type="email"
              placeholder="admin@traefikx.local"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="password">Password</Label>
            <Input
              id="password"
              v-model="password"
              type="password"
              placeholder="••••••••"
              required
            />
          </div>
          <Button
            type="submit"
            class="w-full"
            :disabled="isLoading"
          >
            {{ isLoading ? 'Signing in...' : 'Sign in' }}
          </Button>
        </form>

        <div v-if="oidcEnabled" class="relative">
          <div class="absolute inset-0 flex items-center">
            <span class="w-full border-t" />
          </div>
          <div class="relative flex justify-center text-xs uppercase">
            <span class="bg-card px-2 text-muted-foreground">
              Or continue with
            </span>
          </div>
        </div>

        <Button
          v-if="oidcEnabled"
          variant="outline"
          class="w-full"
          @click="handleOIDCLogin"
        >
          {{ oidcProviderName }}
        </Button>

        <div v-if="error" class="p-3 text-sm text-red-500 bg-red-50 rounded-md">
          {{ error }}
        </div>
      </CardContent>
      <CardFooter>
        <p class="text-sm text-muted-foreground text-center w-full">
          Default: admin@traefikx.local / changeme
        </p>
      </CardFooter>
    </Card>
  </div>
</template>