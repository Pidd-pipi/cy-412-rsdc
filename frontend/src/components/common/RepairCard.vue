<template><article class="card repair-card" :class="{'repair-overdue':repair.overdue}">
  <div class="repair-main">
    <div class="repair-title-line">
      <h3>{{repair.title}}</h3>
      <el-tag :type="urgencyTagType" effect="dark" size="small">{{repair.urgency||'普通'}}</el-tag>
      <el-tag v-if="repair.overdue" type="danger" effect="plain" size="small">响应超时</el-tag>
    </div>
    <p>{{repair.description}}</p>
    <small>{{repair.type}} · {{repair.user?.nickname||'业主'}} · 提交 {{formatDateTime(repair.created_at)}}</small>
    <div class="repair-sla">
      <span>首次响应截止：{{formatDateTime(repair.response_deadline)}}<em v-if="countdown" :class="{'sla-overdue':repair.overdue}">（{{countdown}}）</em></span>
      <span v-if="repair.responded_at">首次响应：{{formatDateTime(repair.responded_at)}} · 耗时 {{formatDuration(repair.response_duration)}}</span>
      <span v-if="repair.handler">处理人：{{repair.handler.nickname}}</span>
    </div>
    <div v-if="isStaff && !ended" class="repair-actions">
      <el-button v-if="!repair.handler_id" type="primary" size="small" :loading="busy===repair.id" @click="emit('accept',repair)">接单</el-button>
      <template v-else>
        <el-button v-if="repair.status==='assigned'" type="primary" size="small" plain @click="emit('progress',{repair,status:'processing'})">开始处理</el-button>
        <el-button v-if="repair.status==='processing'" type="success" size="small" plain @click="emit('progress',{repair,status:'done'})">标记完成</el-button>
      </template>
    </div>
  </div>
  <RepairStatusBadge :status="repair.status"/>
</article></template>
<script setup lang="ts">
import {computed}from'vue';
import type{Repair,RepairStatus}from'../../types';
import {repairUrgencyTagType}from'../../constants/repair';
import {formatDateTime,formatDuration,formatCountdown}from'../../utils/repair';
import RepairStatusBadge from'./RepairStatusBadge.vue';
const p=defineProps<{repair:Repair;isStaff:boolean;busy?:number}>();
const emit=defineEmits<{(e:'accept',repair:Repair):void;(e:'progress',v:{repair:Repair;status:RepairStatus}):void}>();
const urgencyTagType=computed(()=>repairUrgencyTagType[p.repair.urgency||'普通']);
const ended=computed(()=>['done','closed'].includes(p.repair.status));
const countdown=computed(()=>formatCountdown(p.repair));
</script>
