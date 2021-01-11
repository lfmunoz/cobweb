
Vue.component('xInstance', {
    template: `
  <div style="margin: 5px 0px">

 
    <x-modifiable label="Id" v-model="instance.Id"  :readonly="true"> </x-modifiable>
    <x-modifiable label="Address" v-model="instance.Address"  :readonly="false"> </x-modifiable>
    <x-modifiable label="NodeId" v-model="instance.NodeId"  :readonly="false"> </x-modifiable>
    <x-modifiable label="Version" v-model="instance.Version"  :readonly="false"> </x-modifiable>

    <label style="font-size: 0.9rem; display: inline-block;text-align: right;font-weight:900;min-width:100px">Local</label>
    <x-modifiable label="Name" v-model="instance.Local.Name"  :readonly="false"> </x-modifiable>
    <x-modifiable label="Address" v-model="instance.Local.Address"  :readonly="false"> </x-modifiable>
    <x-modifiable label="Port" v-model="instance.Local.Port"  :readonly="false"> </x-modifiable>

    <label style="font-size: 0.9rem; display: inline-block;text-align: right;font-weight:900;min-width:100px">Remote</label>
    <x-modifiable label="Name" v-model="instance.Remote.Name"  :readonly="false"> </x-modifiable>
    <x-modifiable label="Address" v-model="instance.Remote.Address"  :readonly="false"> </x-modifiable>
    <x-modifiable label="Port" v-model="instance.Remote.Port"  :readonly="false"> </x-modifiable>

    <button @click="save">SAVE {{instance.Id}}</button>

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
        }
    },
    mounted() {
        console.log("instance mounted")
        console.log(this.instance)
    }
})
