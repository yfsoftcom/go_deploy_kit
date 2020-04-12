import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';

import { NbSidebarModule, NbLayoutModule, NbButtonModule } from '@nebular/theme';

const routes: Routes = [];

@NgModule({
  imports: [
    RouterModule.forRoot(routes),
    NbLayoutModule,
    NbSidebarModule, // NbSidebarModule.forRoot(), //if this is your app.module
    NbButtonModule,
  ],
  exports: [RouterModule]
})
export class AppRoutingModule { }
