<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import Button from '@/components/ui/button/Button.vue'
import Card from '@/components/ui/card/Card.vue'
import CardHeader from '@/components/ui/card/CardHeader.vue'
import CardTitle from '@/components/ui/card/CardTitle.vue'
import CardDescription from '@/components/ui/card/CardDescription.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import { Users, User, LogOut } from 'lucide-vue-next'

const authStore = useAuthStore()
const router = useRouter()

async function handleLogout() {
  await authStore.logout()
  router.push('/login')
}

function navigateToUsers() {
  router.push('/users')
}

function navigateToProfile() {
  router.push('/profile')
}
</script>

<template>
  <div class="min-h-screen bg-background">
    <header class="border-b">
      <div class="container mx-auto px-4 py-4 flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold">TraefikX</h1>
          <p class="text-sm text-muted-foreground">Welcome, {{ authStore.user?.email }}</p>
        </div>
        <div class="flex items-center gap-4">
          <span 
            v-if="authStore.user?.role === 'admin'"
            class="px-2 py-1 text-xs font-medium bg-primary text-primary-foreground rounded"
          >
            Admin
          </span>
          <Button variant="ghost" @click="handleLogout">
            <LogOut class="w-4 h-4 mr-2" />
            Logout
          </Button>
        </div>
      </div>
    </header>

    <main class="container mx-auto px-4 py-8">
      <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        <Card 
          v-if="authStore.isAdmin" 
          class="cursor-pointer hover:bg-accent transition-colors"
          @click="navigateToUsers"
        >
          <CardHeader>
            <div class="flex items-center gap-2">
              <Users class="w-5 h-5 text-primary" />
              <CardTitle>User Management</CardTitle>
            </div>
            <CardDescription>
              Manage users, reset passwords, and configure access
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button variant="outline" class="w-full">Manage Users</Button>
          </CardContent>
        </Card>

        <Card 
          class="cursor-pointer hover:bg-accent transition-colors"
          @click="navigateToProfile"
        >
          <CardHeader>
            <div class="flex items-center gap-2">
              <User class="w-5 h-5 text-primary" />
              <CardTitle>Profile</CardTitle>
            </div>
            <CardDescription>
              Manage your account, password, and OIDC connections
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button variant="outline" class="w-full">View Profile</Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Proxies</CardTitle>
            <CardDescription>
              Manage Traefik proxies (coming soon)
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button variant="outline" class="w-full" disabled>Coming Soon</Button>
          </CardContent>
        </Card>
      </div>
    </main>
  </div>
</template>