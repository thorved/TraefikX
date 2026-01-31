<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import type { User } from '@/types'
import Button from '@/components/ui/button/Button.vue'
import Input from '@/components/ui/input/Input.vue'
import Label from '@/components/ui/label/Label.vue'
import Card from '@/components/ui/card/Card.vue'
import CardHeader from '@/components/ui/card/CardHeader.vue'
import CardTitle from '@/components/ui/card/CardTitle.vue'
import CardDescription from '@/components/ui/card/CardDescription.vue'
import CardContent from '@/components/ui/card/CardContent.vue'
import { ArrowLeft, Plus, Trash2, RefreshCw, UserX, Check } from 'lucide-vue-next'

const router = useRouter()
const users = ref<User[]>([])
const isLoading = ref(false)
const error = ref('')

// Modal states
const showCreateModal = ref(false)
const showResetPasswordModal = ref(false)
const selectedUser = ref<User | null>(null)
const newPassword = ref('')

// Create user form
const newUser = ref({
  email: '',
  password: '',
  role: 'user' as 'admin' | 'user',
  oidc_enabled: false,
})

onMounted(() => {
  fetchUsers()
})

async function fetchUsers() {
  isLoading.value = true
  try {
    users.value = await api.listUsers()
  } catch (err: any) {
    error.value = 'Failed to load users'
  } finally {
    isLoading.value = false
  }
}

function goBack() {
  router.push('/')
}

async function createUser() {
  try {
    await api.createUser(newUser.value)
    await fetchUsers()
    showCreateModal.value = false
    newUser.value = { email: '', password: '', role: 'user', oidc_enabled: false }
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to create user'
  }
}

async function deleteUser(userId: number) {
  if (!confirm('Are you sure you want to delete this user?')) return
  
  try {
    await api.deleteUser(userId)
    await fetchUsers()
  } catch (err: any) {
    error.value = 'Failed to delete user'
  }
}

async function resetPassword() {
  if (!selectedUser.value) return
  
  try {
    await api.resetPassword(selectedUser.value.id, newPassword.value)
    showResetPasswordModal.value = false
    newPassword.value = ''
    selectedUser.value = null
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to reset password'
  }
}

async function toggleUserActive(user: User) {
  try {
    await api.updateUser(user.id, { is_active: !user.is_active })
    await fetchUsers()
  } catch (err: any) {
    error.value = 'Failed to update user'
  }
}

function formatDate(dateString?: string) {
  if (!dateString) return 'Never'
  return new Date(dateString).toLocaleDateString()
}
</script>

<template>
  <div class="min-h-screen bg-background">
    <header class="border-b">
      <div class="container mx-auto px-4 py-4 flex items-center justify-between">
        <div class="flex items-center gap-4">
          <Button variant="ghost" @click="goBack">
            <ArrowLeft class="w-4 h-4 mr-2" />
            Back
          </Button>
          <h1 class="text-2xl font-bold">User Management</h1>
        </div>
        <Button @click="showCreateModal = true">
          <Plus class="w-4 h-4 mr-2" />
          Add User
        </Button>
      </div>
    </header>

    <main class="container mx-auto px-4 py-8">
      <div v-if="error" class="mb-4 p-4 text-red-500 bg-red-50 rounded-md">
        {{ error }}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Users</CardTitle>
          <CardDescription>Manage user accounts and permissions</CardDescription>
        </CardHeader>
        <CardContent>
          <div v-if="isLoading" class="text-center py-8">Loading...</div>
          
          <table v-else class="w-full">
            <thead>
              <tr class="border-b">
                <th class="text-left py-2 px-4">Email</th>
                <th class="text-left py-2 px-4">Role</th>
                <th class="text-left py-2 px-4">Status</th>
                <th class="text-left py-2 px-4">Auth Methods</th>
                <th class="text-left py-2 px-4">Last Login</th>
                <th class="text-right py-2 px-4">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in users" :key="user.id" class="border-b">
                <td class="py-2 px-4">{{ user.email }}</td>
                <td class="py-2 px-4">
                  <span 
                    :class="user.role === 'admin' ? 'text-primary font-medium' : ''"
                  >
                    {{ user.role }}
                  </span>
                </td>
                <td class="py-2 px-4">
                  <span 
                    :class="user.is_active ? 'text-green-600' : 'text-red-600'"
                  >
                    {{ user.is_active ? 'Active' : 'Inactive' }}
                  </span>
                </td>
                <td class="py-2 px-4">
                  <div class="flex gap-2">
                    <span v-if="user.password_enabled" class="text-xs bg-secondary px-2 py-1 rounded">Password</span>
                    <span v-if="user.oidc_enabled" class="text-xs bg-secondary px-2 py-1 rounded">OIDC</span>
                  </div>
                </td>
                <td class="py-2 px-4">{{ formatDate(user.last_login_at) }}</td>
                <td class="py-2 px-4 text-right">
                  <div class="flex justify-end gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      @click="selectedUser = user; showResetPasswordModal = true"
                    >
                      <RefreshCw class="w-4 h-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      @click="toggleUserActive(user)"
                    >
                      <component :is="user.is_active ? UserX : Check" class="w-4 h-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      @click="deleteUser(user.id)"
                    >
                      <Trash2 class="w-4 h-4 text-destructive" />
                    </Button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </CardContent>
      </Card>
    </main>

    <!-- Create User Modal -->
    <div 
      v-if="showCreateModal" 
      class="fixed inset-0 bg-black/50 flex items-center justify-center p-4"
      @click.self="showCreateModal = false"
    >
      <Card class="w-full max-w-md">
        <CardHeader>
          <CardTitle>Create New User</CardTitle>
          <CardDescription>Add a new user to the system</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <Label>Email</Label>
            <Input v-model="newUser.email" type="email" placeholder="user@example.com" />
          </div>
          
          <div class="space-y-2">
            <Label>Password (optional)</Label>
            <Input v-model="newUser.password" type="password" placeholder="Leave empty for OIDC only" />
          </div>
          
          <div class="space-y-2">
            <Label>Role</Label>
            <select v-model="newUser.role" class="w-full h-10 rounded-md border border-input px-3">
              <option value="user">User</option>
              <option value="admin">Admin</option>
            </select>
          </div>
          
          <div class="flex items-center gap-2">
            <input 
              id="oidc-enabled" 
              v-model="newUser.oidc_enabled" 
              type="checkbox"
              class="rounded border-gray-300"
            />
            <Label for="oidc-enabled">Allow OIDC Login</Label>
          </div>
          
          <div class="flex gap-2">
            <Button variant="outline" class="flex-1" @click="showCreateModal = false">Cancel</Button>
            <Button class="flex-1" @click="createUser">Create User</Button>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Reset Password Modal -->
    <div 
      v-if="showResetPasswordModal" 
      class="fixed inset-0 bg-black/50 flex items-center justify-center p-4"
      @click.self="showResetPasswordModal = false"
    >
      <Card class="w-full max-w-md">
        <CardHeader>
          <CardTitle>Reset Password</CardTitle>
          <CardDescription>
            Set new password for {{ selectedUser?.email }}
          </CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <Label>New Password</Label>
            <Input v-model="newPassword" type="password" placeholder="Min 12 characters" />
          </div>
          
          <div class="flex gap-2">
            <Button variant="outline" class="flex-1" @click="showResetPasswordModal = false">Cancel</Button>
            <Button class="flex-1" @click="resetPassword">Reset Password</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>