<template>
  <section>
    <header class="page-head">
      <div>
        <p class="eyebrow">REPAIR CENTER</p>
        <h2>报修管理</h2>
      </div>
      <el-button type="primary" @click="openCreate">提交报修</el-button>
    </header>

    <div class="toolbar">
      <el-select v-model="status" placeholder="全部状态" clearable style="width:150px" @change="load">
        <el-option v-for="(t,k) in repairStatusText" :key="k" :label="t" :value="k"/>
      </el-select>
      <el-select v-model="urgency" placeholder="全部等级" clearable style="width:140px" @change="load">
        <el-option v-for="u in REPAIR_URGENCY" :key="u" :label="`${u}（${repairUrgencyHours[u]}小时内响应）`" :value="u"/>
      </el-select>
      <el-checkbox v-model="overdueOnly" @change="load">仅看超时工单</el-checkbox>
    </div>

    <div class="repair-list">
      <RepairCard
        v-for="v in filteredItems"
        :key="v.id"
        :repair="v"
        :is-staff="isStaff"
        :busy="busyId"
        @accept="accept"
        @progress="changeStatus"
      />
      <EmptyState v-if="!filteredItems.length"/>
    </div>

    <el-dialog v-model="dialog" title="提交报修工单" width="480px">
      <el-form label-position="top">
        <el-form-item class="repair-form-field" label="报修标题">
          <el-input v-model="form.title" placeholder="报修标题"/>
        </el-form-item>
        <el-form-item class="repair-form-field" label="问题描述">
          <el-input v-model="form.description" type="textarea" placeholder="问题描述"/>
        </el-form-item>
        <el-form-item class="repair-form-field" label="报修类型">
          <el-select v-model="form.type" style="width:100%">
            <el-option label="水电" value="水电"/>
            <el-option label="家具" value="家具"/>
            <el-option label="公共设施" value="公共设施"/>
            <el-option label="其他" value="其他"/>
          </el-select>
        </el-form-item>
        <el-form-item class="repair-form-field" label="紧急等级（决定首次响应时限）">
          <el-radio-group v-model="form.urgency">
            <el-radio v-for="u in REPAIR_URGENCY" :key="u" :value="u">{{u}} · {{repairUrgencyHours[u]}}小时内响应</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog=false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">提交</el-button>
      </template>
    </el-dialog>
  </section>
</template>
<script setup lang="ts">
import {ref,computed,onMounted}from'vue';
import {ElMessage}from'element-plus';
import {listRepairs,createRepair,assignRepair,updateRepairStatus}from'../api/repair';
import {repairStatusText,REPAIR_URGENCY,repairUrgencyHours}from'../constants/repair';
import type{Repair,RepairStatus,RepairUrgency}from'../types';
import {useAuth}from'../hooks/useAuth';
import {authStore}from'../stores/authStore';
import RepairCard from'../components/common/RepairCard.vue';
import EmptyState from'../components/common/EmptyState.vue';

const {isStaff}=useAuth();
const items=ref<Repair[]>([]);
const status=ref(''),urgency=ref<RepairUrgency|''>(''),overdueOnly=ref(false);
const dialog=ref(false),submitting=ref(false),busyId=ref<number>();
const form=ref<{title:string;description:string;type:string;urgency:RepairUrgency;images:string}>({title:'',description:'',type:'水电',urgency:'普通',images:''});

const filteredItems=computed(()=>urgency.value?items.value.filter(v=>(v.urgency||'普通')===urgency.value):items.value);

async function load(){
  items.value=await listRepairs({status:status.value,overdue:overdueOnly.value});
}
function openCreate(){
  form.value={title:'',description:'',type:'水电',urgency:'普通',images:''};
  dialog.value=true;
}
async function submit(){
  try{
    submitting.value=true;
    await createRepair(form.value);
    ElMessage.success('报修工单已提交');
    dialog.value=false;
    await load();
  }catch(e){ElMessage.error((e as Error).message)}finally{submitting.value=false}
}
async function accept(repair:Repair){
  try{
    busyId.value=repair.id;
    const me=authStore.user;
    if(!me)return;
    await assignRepair(repair.id,me.id);
    ElMessage.success('已接单，首次响应时间已记录');
    await load();
  }catch(e){ElMessage.error((e as Error).message)}finally{busyId.value=undefined}
}
async function changeStatus({repair,status:s}:{repair:Repair;status:RepairStatus}){
  try{
    busyId.value=repair.id;
    await updateRepairStatus(repair.id,s);
    ElMessage.success('工单进度已更新');
    await load();
  }catch(e){ElMessage.error((e as Error).message)}finally{busyId.value=undefined}
}
onMounted(load);
</script>
