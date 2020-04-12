import { Component } from '@angular/core';
import { NbMenuItem } from '@nebular/theme';
@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'ui';
  menuItems: NbMenuItem[] = [
    {
      title: 'Dasboard',
      link: '/', // goes into angular `routerLink`
    },
    {
      title: 'Setting',
      children: [
        {
          title: 'Dasboard',
          link: '', // goes into angular `routerLink`
        },
        {
          title: 'Setting',
          url: '/example/menu/menu-link-params.component#some-location', // goes directly into `href` attribute
        },
      ]
    },
  ];
  toggle(){

  }
}
