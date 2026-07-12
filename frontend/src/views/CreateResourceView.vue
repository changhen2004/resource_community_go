<template>
  <el-container class="create-page">
    <el-main class="create-main">
      <section class="create-hero">
        <div>
          <p class="eyebrow">PUBLISH RESOURCE</p>
          <h1>发布一份新的社区资源</h1>
          <p class="hero-copy">
            上传封面和配图，设置标签、可见状态与积分门槛，把内容发布到资源广场。
          </p>
        </div>
        <el-alert
          v-if="!authStore.isAuthenticated"
          type="warning"
          :closable="false"
          show-icon
          title="当前未登录，登录后才能发布资源"
        />
      </section>

      <section v-if="!authStore.isAuthenticated" class="state-panel">
        <el-empty description="请先登录后再发布资源">
          <el-button type="primary" @click="router.push({ name: 'Login' })">前往登录</el-button>
        </el-empty>
      </section>

      <section v-else class="form-shell">
        <el-form label-position="top" class="create-form">
          <div class="form-grid">
            <el-form-item label="标题">
              <el-input v-model="form.title" maxlength="200" show-word-limit placeholder="给资源起一个清晰的标题" />
            </el-form-item>

            <el-form-item label="摘要">
              <el-input
                v-model="form.preview"
                type="textarea"
                :rows="3"
                maxlength="500"
                show-word-limit
                placeholder="用一段摘要说明这份资源的价值"
              />
            </el-form-item>
          </div>

          <el-form-item label="正文内容">
            <el-input
              v-model="form.content"
              type="textarea"
              :rows="12"
              maxlength="20000"
              show-word-limit
              placeholder="输入正文内容，支持在下方补充正文配图"
            />
          </el-form-item>

          <div class="form-grid">
            <el-form-item label="标签">
              <el-input
                v-model="tagsInput"
                placeholder="多个标签用英文逗号分隔，例如 go,backend,notes"
              />
            </el-form-item>

            <el-form-item label="发布状态">
              <el-select v-model="form.status" class="full-width">
                <el-option label="立即发布" value="published" />
                <el-option label="保存草稿" value="draft" />
                <el-option label="归档" value="archived" />
              </el-select>
            </el-form-item>
          </div>

          <div class="form-grid">
            <el-form-item label="访问模式">
              <el-radio-group v-model="accessMode">
                <el-radio-button label="free">免费阅读</el-radio-button>
                <el-radio-button label="paid">积分解锁</el-radio-button>
              </el-radio-group>
            </el-form-item>

            <el-form-item label="所需积分">
              <el-input-number
                v-model="form.requiredPoints"
                class="full-width"
                :min="0"
                :max="10000"
                :disabled="accessMode === 'free'"
              />
            </el-form-item>
          </div>

          <div class="upload-grid">
            <section class="upload-card">
              <div class="upload-card__head">
                <div>
                  <p class="panel-kicker">Cover</p>
                  <h2>封面图</h2>
                </div>
                <el-button :loading="coverUploading" @click="coverInput?.click()">上传封面</el-button>
              </div>
              <input
                ref="coverInput"
                class="hidden-input"
                type="file"
                accept="image/*"
                @change="handleCoverSelected"
              />
              <img v-if="form.coverUrl" :src="form.coverUrl" alt="封面图预览" class="cover-preview" />
              <el-empty v-else description="尚未上传封面图" />
            </section>

            <section class="upload-card">
              <div class="upload-card__head">
                <div>
                  <p class="panel-kicker">Gallery</p>
                  <h2>正文配图</h2>
                </div>
                <el-button :loading="contentImagesUploading" @click="contentImagesInput?.click()">
                  上传配图
                </el-button>
              </div>
              <input
                ref="contentImagesInput"
                class="hidden-input"
                type="file"
                accept="image/*"
                multiple
                @change="handleContentImagesSelected"
              />
              <div v-if="form.contentImages.length" class="gallery-grid">
                <img
                  v-for="(image, index) in form.contentImages"
                  :key="`${image}-${index}`"
                  :src="image"
                  :alt="`正文配图 ${index + 1}`"
                />
              </div>
              <el-empty v-else description="尚未上传正文配图" />
            </section>
          </div>

          <div class="form-actions">
            <el-button @click="resetForm">重置内容</el-button>
            <el-button type="primary" :loading="submitting" @click="handleSubmit">发布资源</el-button>
          </div>
        </el-form>
      </section>
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { createArticle } from '../api/article';
import { uploadContentImages, uploadCover } from '../api/upload';
import { useAuthStore } from '../store/auth';

type AccessMode = 'free' | 'paid';

const router = useRouter();
const authStore = useAuthStore();

const coverInput = ref<HTMLInputElement | null>(null);
const contentImagesInput = ref<HTMLInputElement | null>(null);
const coverUploading = ref(false);
const contentImagesUploading = ref(false);
const submitting = ref(false);
const accessMode = ref<AccessMode>('free');
const tagsInput = ref('');

const createInitialForm = () => ({
  title: '',
  preview: '',
  content: '',
  coverUrl: '',
  contentImages: [] as string[],
  status: 'published' as 'draft' | 'published' | 'archived',
  requiredPoints: 0,
});

const form = ref(createInitialForm());

const normalizedTags = computed(() =>
  tagsInput.value
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean),
);

const resetFileInput = (input: HTMLInputElement | null) => {
  if (input) {
    input.value = '';
  }
};

const handleCoverSelected = async (event: Event) => {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (!file) {
    return;
  }

  coverUploading.value = true;
  try {
    const response = await uploadCover(file);
    form.value.coverUrl = response.url;
    ElMessage.success('封面上传成功');
  } catch (error) {
    console.error('Failed to upload cover:', error);
    ElMessage.error('封面上传失败，请稍后重试。');
  } finally {
    coverUploading.value = false;
    resetFileInput(target);
  }
};

const handleContentImagesSelected = async (event: Event) => {
  const target = event.target as HTMLInputElement;
  const files = Array.from(target.files ?? []);
  if (!files.length) {
    return;
  }

  contentImagesUploading.value = true;
  try {
    const response = await uploadContentImages(files);
    form.value.contentImages = [...form.value.contentImages, ...response.urls];
    ElMessage.success('正文配图上传成功');
  } catch (error) {
    console.error('Failed to upload content images:', error);
    ElMessage.error('正文配图上传失败，请稍后重试。');
  } finally {
    contentImagesUploading.value = false;
    resetFileInput(target);
  }
};

const validateForm = () => {
  if (!form.value.title.trim()) {
    ElMessage.error('标题不能为空');
    return false;
  }
  if (!form.value.preview.trim()) {
    ElMessage.error('摘要不能为空');
    return false;
  }
  if (!form.value.content.trim()) {
    ElMessage.error('正文内容不能为空');
    return false;
  }
  if (accessMode.value === 'paid' && form.value.requiredPoints <= 0) {
    ElMessage.error('积分解锁资源需要设置大于 0 的积分门槛');
    return false;
  }
  return true;
};

const resetForm = () => {
  form.value = createInitialForm();
  tagsInput.value = '';
  accessMode.value = 'free';
  resetFileInput(coverInput.value);
  resetFileInput(contentImagesInput.value);
};

const handleSubmit = async () => {
  if (!authStore.isAuthenticated) {
    ElMessage.error('请先登录后再发布资源');
    router.push({ name: 'Login' });
    return;
  }

  if (!validateForm()) {
    return;
  }

  submitting.value = true;
  try {
    const article = await createArticle({
      title: form.value.title.trim(),
      preview: form.value.preview.trim(),
      content: form.value.content.trim(),
      coverUrl: form.value.coverUrl || undefined,
      contentImages: form.value.contentImages,
      tags: normalizedTags.value,
      status: form.value.status,
      isFree: accessMode.value === 'free',
      requiredPoints: accessMode.value === 'paid' ? form.value.requiredPoints : 0,
    });
    ElMessage.success('资源发布成功');
    router.push({ name: 'ResourceDetail', params: { id: article.id } });
  } catch (error) {
    console.error('Failed to create article:', error);
    ElMessage.error('资源发布失败，请稍后再试。');
  } finally {
    submitting.value = false;
  }
};
</script>

<style scoped>
.create-page {
  min-height: 100%;
  background:
    radial-gradient(circle at top left, rgba(255, 219, 180, 0.42), transparent 24%),
    linear-gradient(180deg, #fbf5ee 0%, #f2e6da 100%);
}

.create-main {
  max-width: 1180px;
  margin: 0 auto;
  padding: 28px 24px 56px;
}

.create-hero,
.state-panel,
.form-shell,
.upload-card {
  border: 1px solid rgba(101, 58, 37, 0.08);
  border-radius: 28px;
  background: rgba(255, 252, 247, 0.9);
  box-shadow: 0 18px 42px rgba(84, 53, 37, 0.08);
  backdrop-filter: blur(14px);
}

.create-hero {
  display: grid;
  gap: 20px;
  padding: 32px;
  margin-bottom: 24px;
}

.eyebrow,
.panel-kicker {
  margin: 0 0 8px;
  color: #935124;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.create-hero h1 {
  margin: 0;
  color: #402514;
  font-size: clamp(30px, 4vw, 48px);
}

.hero-copy {
  max-width: 720px;
  margin: 16px 0 0;
  color: #6b4b37;
  line-height: 1.8;
}

.state-panel,
.form-shell {
  padding: 28px;
}

.create-form {
  display: grid;
  gap: 20px;
}

.form-grid,
.upload-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
}

.full-width {
  width: 100%;
}

.upload-card {
  padding: 24px;
}

.upload-card__head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  margin-bottom: 20px;
}

.upload-card__head h2 {
  margin: 0;
  color: #402514;
}

.cover-preview,
.gallery-grid img {
  width: 100%;
  border-radius: 20px;
  object-fit: cover;
}

.cover-preview {
  max-height: 320px;
}

.gallery-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 12px;
}

.gallery-grid img {
  height: 160px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.hidden-input {
  display: none;
}

@media (max-width: 860px) {
  .form-grid,
  .upload-grid {
    grid-template-columns: 1fr;
  }
}
</style>
