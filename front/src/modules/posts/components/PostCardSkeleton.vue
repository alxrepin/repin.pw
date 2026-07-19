<template>
  <div class="ui-card flex min-h-60 flex-col md:h-60 md:max-h-120 md:flex-row" aria-hidden="true">
    <div class="m-1 shrink-0 md:hidden">
      <div class="skeleton h-40 w-full rounded-3xl" />
    </div>

    <div class="flex min-w-0 flex-1 flex-col justify-between p-5 md:w-[70%]">
      <div>
        <div class="skeleton h-6 w-4/5 rounded-lg" />
        <div class="mt-4 space-y-2.5">
          <div class="skeleton h-3.5 w-full rounded" />
          <div class="skeleton h-3.5 w-full rounded" />
          <div class="skeleton h-3.5 w-2/3 rounded" />
        </div>
      </div>
      <div class="skeleton mt-4 h-3.5 w-24 rounded" />
    </div>

    <div class="m-1 hidden w-[30%] shrink-0 md:block">
      <div class="skeleton h-full w-full rounded-3xl" />
    </div>
  </div>
</template>
