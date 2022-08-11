<template>
  <div id="app">
    <el-menu
      class="el-menu-demo"
      mode="horizontal"
      background-color="#545c64"
      text-color="#fff"
      active-text-color="#ffd04b"
    >
      <el-menu-item index="1"><router-link to="/"> Льготники</router-link> </el-menu-item>

<!--      <el-menu-item index="2"-->
<!--        ><router-link to="/retirees"> Льготники</router-link>-->
<!--      </el-menu-item>-->
<!--      <el-menu-item index="3"-->
<!--        ><router-link to="/breakers"> Нарушители</router-link>-->
<!--      </el-menu-item>-->

      <el-menu-item index="4"
        ><router-link to="/updates"> Обновления</router-link>
      </el-menu-item>
    </el-menu>

    <div class="container mt-3">
      <router-view />
    </div>
  </div>
</template>

<script>
export default {
  name: "app",
};
</script>
