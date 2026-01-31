<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api/client'
import Card from '@/components/ui/card/Card.vue'
import CardHeader from '@/components/ui/card/CardHeader.vue'
import CardTitle from '@/components/ui/card/CardTitle.vue'
import CardDescription from '@/components/ui/card/CardDescription.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import Button from '@/components/ui/button/Button.vue'
import { Loader2, AlertCircle, CheckCircle } from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isLoading = ref(true)
const error = ref('')
const success = ref(false)

onMounted(async () => {
  const code = route.query.code as string
  const state = route.query.state as string

  if (!code || !state) {
    error.value = 'Invalid callback parameters'
    isLoading.value = false
    return
  }

  try {
    const response = await api.handleOIDCCallback(code, state)
    authStore.user = response.user
    success.value = true
    
    // Redirect to dashboard after a short delay
    setTimeout(() => {
      router.push('/')
    }, 2000)
  } catch (err: any) {
    error.value = err.response?.data?.error || 'OIDC authentication failed'
  } finally {
    isLoading.value = false
  }
})
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-background p-4">
    <Card class="w-full max-w-md">
      <CardHeader class="text-center">
        <CardTitle class="text-2xl">OIDC Authentication</CardTitle>
        <CardDescription>
          {{ isLoading ? 'Processing authentication...' : error ? 'Authentication failed' : 'Authentication successful' }}
        </CardDescription>
      </CardHeader>
      <CardContent class="flex flex-col items-center space-y-4">
        <div v-if="isLoading" class="py-8">
          <Loader2 class="w-12 h-12 animate-spin text-primary" />
        </div>

        <div v-else-if="error" class="text-center space-y-4">
          <AlertCircle class="w-12 h-12 text-destructive mx-auto" />
          <p class="text-destructive">{{ error }}</p>
          <Button @click="router.push('/login')">Back to Login</Button>
        </div>

        <div v-else-if="success" class="text-center space-y-4">
          <CheckCircle class="w-12 h-12 text-green-600 mx-auto" />
          <p class="text-green-600">Successfully authenticated!</p>
          <p class="text-sm text-muted-foreground">Redirecting to dashboard...</p>
        </div>
      </CardContent>
    </Card>
  </div>
</template>