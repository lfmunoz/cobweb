
Vue.component('xInstance', {
    template: `
  <div style="margin: 5px 0px; border: 1px solid black;">


    <h3 style="font-weight:900; margin: 5px; border: 0px; ">{{instance.NodeId}} - v{{instance.Version}}</h3>
 
    <x-modifiable label="Id" v-model="instance.Id"  :readonly="true"> </x-modifiable>
    <x-modifiable label="Address" v-model="instance.Address"  :readonly="false"> </x-modifiable>


    <hr>
    <button @click="add" style="margin: 5px">ADD</button>

    <div v-for="(obj, index) in instance.Local" :key="index">
      <button @click="remove" style="margin: 5px">DELETE</button>
      <div class="two-column">
        <div class="column-left">
          <div class="inst-conf-title">Local listen configuration</div>
          <x-modifiable label="Name" v-model="instance.Local[index].Name"  :readonly="false"> </x-modifiable>
          <x-modifiable label="Address" v-model="instance.Local[index].Address"  :readonly="false"> </x-modifiable>
          <x-modifiable label="Port" v-model="instance.Local[index].Port"  :readonly="false"> </x-modifiable>
        </div>

        <div class="column-right"">
            <div class="inst-conf-title">Remote route configuration</div>
            <x-modifiable label="Name" v-model="instance.Remote[index].Name"  :readonly="false"> </x-modifiable>
            <x-modifiable label="Address" v-model="instance.Remote[index].Address"  :readonly="false"> </x-modifiable>
            <x-modifiable label="Port" v-model="instance.Remote[index].Port"  :readonly="false"> </x-modifiable>
        </div>
      </div>

    </div> 


    <button @click="save" style="margin: 5px">SAVE</button>

  </div>
  `,
    props: ['value', 'test'], // value is default for v-model
    data() {
        return {
            instance: this.value
        }
    },
    methods: {
        save() {
            this.$emit('save', this.instance)
        },
        remove() {
          console.log("remove")

        },
        add() {
          console.log("add")
          this.instance.Local.push({
            Name: "local_new",
            Port: "2000",
            Address: "0.0.0.0",
          })
          this.instance.Remote.push({
            Name: "remote_new",
            Port: "80",
            Address: "apache.org",
          })
        }
    },
    mounted() {
        console.log("instance mounted")
        console.log(this.instance)
        console.log(this.instance.Local)
        console.log(this.instance.Local.length)
        console.log(this.instance.Local[0].Name)
    }
})
