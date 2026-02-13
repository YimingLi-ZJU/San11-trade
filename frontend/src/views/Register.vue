<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <el-icon size="48" color="#667eea"><Trophy /></el-icon>
        <h1>注册账号</h1>
        <p>加入三国志11交易联赛</p>
      </div>
      
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        @submit.prevent="handleRegister"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="form.username"
            prefix-icon="User"
            placeholder="请输入用户名"
            size="large"
          />
        </el-form-item>
        
        <el-form-item label="昵称" prop="nickname">
          <el-input
            v-model="form.nickname"
            prefix-icon="Postcard"
            placeholder="请输入昵称（可选）"
            size="large"
          />
        </el-form-item>
        
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            prefix-icon="Lock"
            placeholder="请输入密码（至少6位）"
            size="large"
            show-password
          />
        </el-form-item>
        
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="form.confirmPassword"
            type="password"
            prefix-icon="Lock"
            placeholder="请再次输入密码"
            size="large"
            show-password
          />
        </el-form-item>

        <!-- Invite code field (shown when required) -->
        <el-form-item 
          v-if="requireInviteCode" 
          label="邀请码" 
          prop="inviteCode"
        >
          <el-input
            v-model="form.inviteCode"
            prefix-icon="Ticket"
            placeholder="请输入邀请码"
            size="large"
            @blur="validateInviteCode"
          >
            <template #append>
              <el-icon v-if="inviteCodeStatus === 'valid'" color="#67c23a"><CircleCheck /></el-icon>
              <el-icon v-else-if="inviteCodeStatus === 'invalid'" color="#f56c6c"><CircleClose /></el-icon>
              <el-icon v-else-if="inviteCodeStatus === 'checking'" class="is-loading"><Loading /></el-icon>
            </template>
          </el-input>
          <div v-if="inviteCodeMessage" :class="['invite-code-msg', inviteCodeStatus]">
            {{ inviteCodeMessage }}
          </div>
        </el-form-item>
        
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            native-type="submit"
            style="width: 100%"
          >
            注册
          </el-button>
        </el-form-item>
      </el-form>
      
      <div class="login-footer">
        <span>已有账号？</span>
        <router-link to="/login">立即登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { CircleCheck, CircleClose, Loading } from '@element-plus/icons-vue'
import { useUserStore } from '../stores/user'
import { gameApi, inviteCodeApi } from '../api'

const router = useRouter()
const userStore = useUserStore()

const formRef = ref(null)
const loading = ref(false)
const requireInviteCode = ref(false)
const inviteCodeStatus = ref('') // '', 'checking', 'valid', 'invalid'
const inviteCodeMessage = ref('')

const form = reactive({
  username: '',
  nickname: '',
  password: '',
  confirmPassword: '',
  inviteCode: ''
})

// Load registration config on mount
onMounted(async () => {
  try {
    const response = await gameApi.getRegistrationConfig()
    requireInviteCode.value = response.data.require_invite_code
  } catch (error) {
    console.error('Failed to load registration config:', error)
    // Default to requiring invite code if config fails
    requireInviteCode.value = true
  }
})

const validateConfirmPassword = (rule, value, callback) => {
  if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const validateInviteCodeField = (rule, value, callback) => {
  if (requireInviteCode.value && !value) {
    callback(new Error('请输入邀请码'))
  } else if (inviteCodeStatus.value === 'invalid') {
    callback(new Error(inviteCodeMessage.value || '邀请码无效'))
  } else {
    callback()
  }
}

const rules = computed(() => ({
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度在3-50个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ],
  inviteCode: requireInviteCode.value ? [
    { required: true, message: '请输入邀请码', trigger: 'blur' },
    { validator: validateInviteCodeField, trigger: 'blur' }
  ] : []
}))

// Validate invite code in real-time
async function validateInviteCode() {
  if (!form.inviteCode || !requireInviteCode.value) {
    inviteCodeStatus.value = ''
    inviteCodeMessage.value = ''
    return
  }

  inviteCodeStatus.value = 'checking'
  inviteCodeMessage.value = ''

  try {
    const response = await inviteCodeApi.validate(form.inviteCode)
    if (response.data.valid) {
      inviteCodeStatus.value = 'valid'
      inviteCodeMessage.value = response.data.remaining > 1 
        ? `有效，剩余 ${response.data.remaining} 次使用` 
        : '有效'
    } else {
      inviteCodeStatus.value = 'invalid'
      inviteCodeMessage.value = response.data.reason || '邀请码无效'
    }
  } catch (error) {
    inviteCodeStatus.value = 'invalid'
    inviteCodeMessage.value = '验证失败，请稍后重试'
  }
}

async function handleRegister() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  // Double-check invite code if required
  if (requireInviteCode.value && inviteCodeStatus.value !== 'valid') {
    await validateInviteCode()
    if (inviteCodeStatus.value !== 'valid') {
      ElMessage.error(inviteCodeMessage.value || '请输入有效的邀请码')
      return
    }
  }
  
  loading.value = true
  try {
    await userStore.register({
      username: form.username,
      password: form.password,
      nickname: form.nickname || form.username,
      invite_code: requireInviteCode.value ? form.inviteCode : ''
    })
    ElMessage.success('注册成功')
    router.push('/')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  padding: 40px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.login-header h1 {
  margin: 16px 0 8px;
  font-size: 24px;
  color: #333;
}

.login-header p {
  color: #999;
  font-size: 14px;
}

.login-footer {
  text-align: center;
  margin-top: 20px;
  color: #666;
}

.login-footer a {
  color: #667eea;
  text-decoration: none;
  margin-left: 5px;
}

.login-footer a:hover {
  text-decoration: underline;
}

.invite-code-msg {
  font-size: 12px;
  margin-top: 4px;
}

.invite-code-msg.valid {
  color: #67c23a;
}

.invite-code-msg.invalid {
  color: #f56c6c;
}

.invite-code-msg.checking {
  color: #909399;
}

.is-loading {
  animation: rotating 2s linear infinite;
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
